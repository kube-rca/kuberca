import React, { useCallback, useEffect, useMemo, useRef, useState, type KeyboardEvent } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import {
  createFeedbackComment,
  deleteFeedbackComment,
  fetchFeedbackSummary,
  FeedbackCommentResponse,
  updateFeedbackComment,
  voteFeedback,
} from '../utils/api';

type VoteType = 'up' | 'down' | null;
type EditorTab = 'write' | 'preview';

interface FeedbackSectionProps {
  targetType: 'incident' | 'alert';
  targetId: string;
}

const avatarColorByName = (name: string) => {
  const palette = [
    'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/40 dark:text-cyan-300',
    'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300',
    'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
    'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300',
    'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-300',
  ];
  const sum = name.split('').reduce((acc, ch) => acc + ch.charCodeAt(0), 0);
  return palette[sum % palette.length];
};

const formatCommentTimestamp = (value: string): string => {
  if (!value) return '';
  const normalized = value.replace('T', ' ');
  if (normalized.length >= 16) {
    return normalized.slice(0, 16);
  }
  return normalized;
};

const useEditorHistory = (initialValue = '') => {
  const [value, setValue] = useState(initialValue);
  const historyRef = useRef<string[]>([initialValue]);
  const indexRef = useRef(0);

  const commit = useCallback((next: string) => {
    const current = historyRef.current[indexRef.current];
    if (next === current) {
      setValue(next);
      return;
    }
    const nextHistory = historyRef.current.slice(0, indexRef.current + 1);
    nextHistory.push(next);
    historyRef.current = nextHistory;
    indexRef.current = nextHistory.length - 1;
    setValue(next);
  }, []);

  const reset = useCallback((next: string) => {
    historyRef.current = [next];
    indexRef.current = 0;
    setValue(next);
  }, []);

  const undo = useCallback(() => {
    if (indexRef.current === 0) {
      return false;
    }
    indexRef.current -= 1;
    setValue(historyRef.current[indexRef.current]);
    return true;
  }, []);

  const redo = useCallback(() => {
    if (indexRef.current >= historyRef.current.length - 1) {
      return false;
    }
    indexRef.current += 1;
    setValue(historyRef.current[indexRef.current]);
    return true;
  }, []);

  return { value, setValue: commit, reset, undo, redo };
};

const FeedbackSection: React.FC<FeedbackSectionProps> = ({ targetType, targetId }) => {
  const [selectedVote, setSelectedVote] = useState<VoteType>(null);
  const [upVotes, setUpVotes] = useState(0);
  const [downVotes, setDownVotes] = useState(0);
  const [comments, setComments] = useState<FeedbackCommentResponse[]>([]);
  const [tab, setTab] = useState<EditorTab>('write');
  const draftEditor = useEditorHistory('');
  const draft = draftEditor.value;
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [moreMenuOpen, setMoreMenuOpen] = useState(false);
  const [commentMenuId, setCommentMenuId] = useState<number | null>(null);
  const [editingCommentId, setEditingCommentId] = useState<number | null>(null);
  const editingDraftEditor = useEditorHistory('');
  const editingDraft = editingDraftEditor.value;
  const [editingTab, setEditingTab] = useState<EditorTab>('write');
  const [editingMoreMenuOpen, setEditingMoreMenuOpen] = useState(false);
  const [commentActionLoadingId, setCommentActionLoadingId] = useState<number | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);
  const editingTextareaRef = useRef<HTMLTextAreaElement | null>(null);
  const moreMenuRef = useRef<HTMLDivElement | null>(null);
  const editingMoreMenuRef = useRef<HTMLDivElement | null>(null);

  const loadFeedback = async () => {
    setLoading(true);
    try {
      const summary = await fetchFeedbackSummary(targetType, targetId);
      setSelectedVote(summary.my_vote ?? null);
      setUpVotes(summary.up_votes ?? 0);
      setDownVotes(summary.down_votes ?? 0);
      setComments(summary.comments ?? []);
    } catch (error) {
      console.error('Failed to load feedback:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFeedback();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [targetType, targetId]);

  useEffect(() => {
    const handleOutsideClick = (event: MouseEvent) => {
      const target = event.target as HTMLElement;

      if (moreMenuRef.current && !moreMenuRef.current.contains(target)) {
        setMoreMenuOpen(false);
      }
      if (editingMoreMenuRef.current && !editingMoreMenuRef.current.contains(target)) {
        setEditingMoreMenuOpen(false);
      }

      if (!target.closest('[data-comment-menu]')) {
        setCommentMenuId(null);
      }
    };

    document.addEventListener('mousedown', handleOutsideClick);
    return () => {
      document.removeEventListener('mousedown', handleOutsideClick);
    };
  }, []);

  const handleVote = async (vote: Exclude<VoteType, null>) => {
    try {
      const nextVote = selectedVote === vote ? 'none' : vote;
      await voteFeedback(targetType, targetId, nextVote);
      await loadFeedback();
    } catch (error) {
      console.error('Failed to save vote:', error);
      alert('Failed to save the vote.');
    }
  };

  const handleSubmit = async () => {
    const body = draft.trim();
    if (!body) return;

    setSubmitting(true);
    try {
      const newComment = await createFeedbackComment(targetType, targetId, body);
      setComments((prev) => [...prev, newComment]);
      draftEditor.reset('');
      setTab('write');
    } catch (error) {
      console.error('Failed to save comment:', error);
      alert('Failed to save the comment.');
    } finally {
      setSubmitting(false);
    }
  };

  const title = useMemo(() => (targetType === 'incident' ? 'Incident Feedback' : 'Alert Feedback'), [targetType]);

  const startEditComment = (comment: FeedbackCommentResponse) => {
    setCommentMenuId(null);
    setEditingCommentId(comment.comment_id);
    editingDraftEditor.reset(comment.body);
    setEditingTab('write');
  };

  const cancelEditComment = () => {
    setEditingCommentId(null);
    editingDraftEditor.reset('');
  };

  const saveEditComment = async (commentId: number) => {
    const body = editingDraft.trim();
    if (!body) {
      alert('Please enter a comment.');
      return;
    }

    setCommentActionLoadingId(commentId);
    try {
      const updated = await updateFeedbackComment(targetType, targetId, commentId, body);
      setComments((prev) => prev.map((item) => (item.comment_id === commentId ? updated : item)));
      cancelEditComment();
    } catch (error) {
      console.error('Failed to update comment:', error);
      alert('Failed to modify the comment.');
    } finally {
      setCommentActionLoadingId(null);
    }
  };

  const removeComment = async (commentId: number) => {
    if (!window.confirm('Are you sure you want to delete the comment?')) return;

    setCommentActionLoadingId(commentId);
    try {
      await deleteFeedbackComment(targetType, targetId, commentId);
      setComments((prev) => prev.filter((item) => item.comment_id !== commentId));
      if (editingCommentId === commentId) {
        cancelEditComment();
      }
    } catch (error) {
      console.error('Failed to delete comment:', error);
      alert('Failed to delete the comment.');
    } finally {
      setCommentActionLoadingId(null);
      setCommentMenuId(null);
    }
  };

  type EditorContext = {
    draft: string;
    setDraft: (value: string) => void;
    undo: () => boolean;
    redo: () => boolean;
    textareaRef: React.RefObject<HTMLTextAreaElement | null>;
  };

  const applySelectionTransform = (
    transform: (selected: string) => { text: string; cursorOffset?: number },
    ctx: EditorContext = {
      draft,
      setDraft: draftEditor.setValue,
      undo: draftEditor.undo,
      redo: draftEditor.redo,
      textareaRef,
    }
  ) => {
    const textarea = ctx.textareaRef.current;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selected = ctx.draft.slice(start, end);
    const next = transform(selected);
    const nextPos = start + (next.cursorOffset ?? next.text.length);

    textarea.focus();
    textarea.setRangeText(next.text, start, end, 'end');
    ctx.setDraft(textarea.value);
    textarea.setSelectionRange(nextPos, nextPos);
  };

  const applyLinePrefix = (
    prefix: string,
    ctx: EditorContext = {
      draft,
      setDraft: draftEditor.setValue,
      undo: draftEditor.undo,
      redo: draftEditor.redo,
      textareaRef,
    }
  ) => {
    const textarea = ctx.textareaRef.current;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const before = ctx.draft.slice(0, start);
    const selected = ctx.draft.slice(start, end);
    const lineStart = before.lastIndexOf('\n') + 1;
    const block = `${ctx.draft.slice(lineStart, start)}${selected}`;
    const prefixed = block
      .split('\n')
      .map((line) => (line.trim() ? `${prefix}${line}` : line || prefix))
      .join('\n');

    textarea.focus();
    textarea.setRangeText(prefixed, lineStart, end, 'end');
    ctx.setDraft(textarea.value);
  };

  const handleEditorKeyDown = (
    event: KeyboardEvent<HTMLTextAreaElement>,
    ctx: EditorContext
  ) => {
    const isMod = event.metaKey || event.ctrlKey;
    if (!isMod) {
      return;
    }

    const key = event.key.toLowerCase();

    if (key === 'z') {
      event.preventDefault();
      if (event.shiftKey) {
        ctx.redo();
      } else {
        ctx.undo();
      }
      return;
    }

    if (key === 'y') {
      event.preventDefault();
      ctx.redo();
      return;
    }

    if (key === 'b') {
      event.preventDefault();
      applySelectionTransform((s) => ({ text: `**${s || 'bold text'}**`, cursorOffset: s ? undefined : 2 }), ctx);
      return;
    }

    if (key === 'i') {
      event.preventDefault();
      applySelectionTransform((s) => ({ text: `*${s || 'italic text'}*`, cursorOffset: s ? undefined : 1 }), ctx);
      return;
    }

    if (key === 'k') {
      event.preventDefault();
      applySelectionTransform(
        (s) => ({ text: `[${s || 'link text'}](https://example.com)`, cursorOffset: s ? undefined : 1 }),
        ctx
      );
      return;
    }

    if (key === 'e') {
      event.preventDefault();
      applySelectionTransform((s) => ({ text: `\`${s || 'code'}\``, cursorOffset: s ? undefined : 1 }), ctx);
      return;
    }

    if (event.shiftKey && key === 'h') {
      event.preventDefault();
      applyLinePrefix('### ', ctx);
    }
  };

  const handleMoreAction = (
    action: 'unordered' | 'numbered' | 'task' | 'mention' | 'reference' | 'slash',
    ctx?: EditorContext
  ) => {
    setMoreMenuOpen(false);
    setEditingMoreMenuOpen(false);

    const editCtx: EditorContext = ctx ?? {
      draft,
      setDraft: draftEditor.setValue,
      undo: draftEditor.undo,
      redo: draftEditor.redo,
      textareaRef,
    };
    if (action === 'unordered') {
      applyLinePrefix('- ', editCtx);
      return;
    }
    if (action === 'numbered') {
      applyLinePrefix('1. ', editCtx);
      return;
    }
    if (action === 'task') {
      applyLinePrefix('- [ ] ', editCtx);
      return;
    }
    if (action === 'mention') {
      applySelectionTransform((s) => ({ text: s ? `@${s}` : '@mention' }), editCtx);
      return;
    }
    if (action === 'reference') {
      applySelectionTransform((s) => ({ text: s ? `${s}#123` : 'owner/repo#123' }), editCtx);
      return;
    }
    applySelectionTransform((s) => ({ text: s ? `/${s}` : '/command' }), editCtx);
  };

  const clearEditor = (ctx: EditorContext) => {
    ctx.setDraft('');
    setTimeout(() => {
      const textarea = ctx.textareaRef.current;
      if (!textarea) return;
      textarea.focus();
      textarea.setSelectionRange(0, 0);
    }, 0);
  };

  return (
    <section className="border-t border-slate-200 dark:border-slate-700 pt-6 mt-2">
      <h3 className="text-lg font-bold text-slate-800 dark:text-slate-100 mb-4">{title}</h3>

      <div className="mb-6 flex items-center justify-center gap-4">
        <button
          type="button"
          onClick={() => handleVote('up')}
          className={`h-12 w-12 rounded-full border text-2xl flex items-center justify-center transition-colors ${
            selectedVote === 'up'
              ? 'border-emerald-500 bg-emerald-50 dark:bg-emerald-900/20'
              : 'border-slate-300 dark:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-700'
          }`}
          aria-label="Upvote"
        >
          👍🏻
        </button>
        <span className="text-sm font-medium text-slate-700 dark:text-slate-300">{upVotes}</span>

        <button
          type="button"
          onClick={() => handleVote('down')}
          className={`h-12 w-12 rounded-full border text-2xl flex items-center justify-center transition-colors ${
            selectedVote === 'down'
              ? 'border-rose-500 bg-rose-50 dark:bg-rose-900/20'
              : 'border-slate-300 dark:border-slate-600 hover:bg-slate-50 dark:hover:bg-slate-700'
          }`}
          aria-label="Downvote"
        >
          👎🏻
        </button>
        <span className="text-sm font-medium text-slate-700 dark:text-slate-300">{downVotes}</span>
      </div>

      <div className="space-y-4 mb-6">
        {loading && (
          <div className="text-sm text-slate-500 dark:text-slate-400">Loading feedback...</div>
        )}
        {comments.map((comment) => {
          const initial = comment.author_login_id[0]?.toUpperCase() || 'U';
          return (
            <article key={comment.comment_id} className="relative border border-slate-200 dark:border-slate-700 rounded-lg overflow-visible">
              <header className="px-4 py-3 bg-slate-50 dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 flex items-center gap-3">
                <div className={`h-8 w-8 rounded-full flex items-center justify-center text-xs font-bold ${avatarColorByName(comment.author_login_id)}`}>
                  {initial}
                </div>
                <div className="text-sm flex-1">
                  <span className="font-semibold text-slate-900 dark:text-slate-100">{comment.author_login_id}</span>
                  <span className="text-slate-500 dark:text-slate-400 ml-2">{formatCommentTimestamp(comment.created_at)}</span>
                </div>
                <div className="relative" data-comment-menu="true">
                  <button
                    type="button"
                    onClick={() => setCommentMenuId((prev) => (prev === comment.comment_id ? null : comment.comment_id))}
                    className="h-7 w-7 rounded text-slate-500 hover:bg-slate-200 dark:text-slate-300 dark:hover:bg-slate-700"
                    aria-label="Comment actions"
                  >
                    ...
                  </button>
                  {commentMenuId === comment.comment_id && (
                    <div className="absolute right-0 top-8 z-[100] w-28 overflow-hidden rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 shadow-lg">
                      <button
                        type="button"
                        onClick={() => startEditComment(comment)}
                        className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800"
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        onClick={() => removeComment(comment.comment_id)}
                        className="w-full px-3 py-2 text-left text-sm text-rose-600 hover:bg-rose-50 dark:text-rose-400 dark:hover:bg-rose-900/20"
                        disabled={commentActionLoadingId === comment.comment_id}
                      >
                        Delete
                      </button>
                    </div>
                  )}
                </div>
              </header>
              <div className="px-4 py-4 text-sm text-slate-800 dark:text-slate-200">
                {editingCommentId === comment.comment_id ? (
                  <div className="border border-slate-200 dark:border-slate-700 rounded-lg overflow-visible -mx-1 -my-1">
                    <div className="flex border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800">
                      <button
                        type="button"
                        onClick={() => setEditingTab('write')}
                        className={`px-4 py-2 text-sm font-medium border-r border-slate-200 dark:border-slate-700 ${
                          editingTab === 'write' ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-white' : 'text-slate-500 dark:text-slate-400'
                        }`}
                      >
                        Write
                      </button>
                      <button
                        type="button"
                        onClick={() => setEditingTab('preview')}
                        className={`px-4 py-2 text-sm font-medium ${
                          editingTab === 'preview' ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-white' : 'text-slate-500 dark:text-slate-400'
                        }`}
                      >
                        Preview
                      </button>
                    </div>
                    {editingTab === 'write' && (
                      <div className="flex items-center gap-1 px-2 py-2 border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 overflow-x-auto">
                        <button
                          type="button"
                          onClick={() =>
                            applyLinePrefix('### ', {
                              draft: editingDraft,
                              setDraft: editingDraftEditor.setValue,
                              undo: editingDraftEditor.undo,
                              redo: editingDraftEditor.redo,
                              textareaRef: editingTextareaRef,
                            })
                          }
                          className="h-8 w-8 rounded text-lg font-semibold text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
                          title="Heading"
                        >
                          H
                        </button>
                        <button
                          type="button"
                          onClick={() =>
                            applySelectionTransform(
                              (s) => ({ text: `**${s || 'bold text'}**`, cursorOffset: s ? undefined : 2 }),
                              {
                                draft: editingDraft,
                                setDraft: editingDraftEditor.setValue,
                                undo: editingDraftEditor.undo,
                                redo: editingDraftEditor.redo,
                                textareaRef: editingTextareaRef,
                              }
                            )
                          }
                          className="h-8 w-8 rounded text-lg font-bold text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
                          title="Bold"
                        >
                          B
                        </button>
                        <button
                          type="button"
                          onClick={() =>
                            applySelectionTransform(
                              (s) => ({ text: `*${s || 'italic text'}*`, cursorOffset: s ? undefined : 1 }),
                              {
                                draft: editingDraft,
                                setDraft: editingDraftEditor.setValue,
                                undo: editingDraftEditor.undo,
                                redo: editingDraftEditor.redo,
                                textareaRef: editingTextareaRef,
                              }
                            )
                          }
                          className="h-8 w-8 rounded text-lg italic text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
                          title="Italic"
                        >
                          I
                        </button>
                        <button
                          type="button"
                          onClick={() =>
                            applySelectionTransform(
                              (s) => ({ text: `\`${s || 'code'}\``, cursorOffset: s ? undefined : 1 }),
                              {
                                draft: editingDraft,
                                setDraft: editingDraftEditor.setValue,
                                undo: editingDraftEditor.undo,
                                redo: editingDraftEditor.redo,
                                textareaRef: editingTextareaRef,
                              }
                            )
                          }
                          className="h-8 w-8 rounded text-lg text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
                          title="Code"
                        >
                          &lt;&gt;
                        </button>
                        <button
                          type="button"
                          onClick={() =>
                            applySelectionTransform(
                              (s) => ({ text: `[${s || 'link text'}](https://example.com)`, cursorOffset: s ? undefined : 1 }),
                              {
                                draft: editingDraft,
                                setDraft: editingDraftEditor.setValue,
                                undo: editingDraftEditor.undo,
                                redo: editingDraftEditor.redo,
                                textareaRef: editingTextareaRef,
                              }
                            )
                          }
                          className="h-8 w-8 rounded text-lg text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
                          title="Link"
                        >
                          🔗
                        </button>
                        <span className="mx-1 h-6 w-px bg-slate-300 dark:bg-slate-600" />
                        <div className="relative" ref={editingMoreMenuRef}>
                          <button
                            type="button"
                            onClick={() => setEditingMoreMenuOpen((v) => !v)}
                            className="h-8 w-8 rounded text-lg text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
                            title="More"
                          >
                            ...
                          </button>
                          {editingMoreMenuOpen && (
                            <div className="absolute right-0 top-9 z-[100] w-52 overflow-hidden rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 shadow-lg">
                              <button type="button" onClick={() => handleMoreAction('unordered', { draft: editingDraft, setDraft: editingDraftEditor.setValue, undo: editingDraftEditor.undo, redo: editingDraftEditor.redo, textareaRef: editingTextareaRef })} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Unordered list</button>
                              <button type="button" onClick={() => handleMoreAction('numbered', { draft: editingDraft, setDraft: editingDraftEditor.setValue, undo: editingDraftEditor.undo, redo: editingDraftEditor.redo, textareaRef: editingTextareaRef })} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Numbered list</button>
                              <button type="button" onClick={() => handleMoreAction('task', { draft: editingDraft, setDraft: editingDraftEditor.setValue, undo: editingDraftEditor.undo, redo: editingDraftEditor.redo, textareaRef: editingTextareaRef })} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Task list</button>
                              <div className="h-px bg-slate-200 dark:bg-slate-700" />
                              <button type="button" onClick={() => handleMoreAction('mention', { draft: editingDraft, setDraft: editingDraftEditor.setValue, undo: editingDraftEditor.undo, redo: editingDraftEditor.redo, textareaRef: editingTextareaRef })} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">@ Mention</button>
                              <button type="button" onClick={() => handleMoreAction('reference', { draft: editingDraft, setDraft: editingDraftEditor.setValue, undo: editingDraftEditor.undo, redo: editingDraftEditor.redo, textareaRef: editingTextareaRef })} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Reference</button>
                              <button type="button" onClick={() => handleMoreAction('slash', { draft: editingDraft, setDraft: editingDraftEditor.setValue, undo: editingDraftEditor.undo, redo: editingDraftEditor.redo, textareaRef: editingTextareaRef })} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Slash commands</button>
                            </div>
                          )}
                        </div>
                        <button
                          type="button"
                          onClick={() =>
                            clearEditor({
                              draft: editingDraft,
                              setDraft: editingDraftEditor.setValue,
                              undo: editingDraftEditor.undo,
                              redo: editingDraftEditor.redo,
                              textareaRef: editingTextareaRef,
                            })
                          }
                          className="ml-auto h-8 w-8 rounded text-base text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
                          title="Clear"
                          aria-label="Clear editor"
                        >
                          ⌫
                        </button>
                      </div>
                    )}
                    {editingTab === 'write' ? (
                      <textarea
                        ref={editingTextareaRef}
                        value={editingDraft}
                        onChange={(e) => editingDraftEditor.setValue(e.target.value)}
                        onKeyDown={(e) =>
                          handleEditorKeyDown(e, {
                            draft: editingDraft,
                            setDraft: editingDraftEditor.setValue,
                            undo: editingDraftEditor.undo,
                            redo: editingDraftEditor.redo,
                            textareaRef: editingTextareaRef,
                          })
                        }
                        className="w-full min-h-[160px] p-4 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none text-sm"
                      />
                    ) : (
                      <div className="min-h-[160px] p-4 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200">
                        {editingDraft.trim() ? (
                          <ReactMarkdown
                            remarkPlugins={[remarkGfm]}
                            components={{
                              p: ({ node: _node, ...props }) => <p className="mb-3 last:mb-0 leading-relaxed" {...props} />,
                              ul: ({ node: _node, ...props }) => <ul className="list-disc pl-5 mb-3 space-y-1" {...props} />,
                              code: ({ node: _node, ...props }) => (
                                <code className="bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded px-1 py-0.5 font-mono text-xs" {...props} />
                              ),
                            }}
                          >
                            {editingDraft}
                          </ReactMarkdown>
                        ) : (
                          <p className="text-sm text-slate-500 dark:text-slate-400">Nothing to preview.</p>
                        )}
                      </div>
                    )}
                    <div className="flex justify-end gap-2 px-4 py-3 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800">
                      <button
                        type="button"
                        onClick={cancelEditComment}
                        className="px-3 py-1.5 rounded border border-slate-300 dark:border-slate-600 text-slate-700 dark:text-slate-200 text-xs font-semibold hover:bg-slate-50 dark:hover:bg-slate-700"
                      >
                        Cancel
                      </button>
                      <button
                        type="button"
                        onClick={() => saveEditComment(comment.comment_id)}
                        disabled={commentActionLoadingId === comment.comment_id}
                        className="px-3 py-1.5 rounded bg-cyan-600 text-white text-xs font-semibold hover:bg-cyan-700 disabled:opacity-50"
                      >
                        {commentActionLoadingId === comment.comment_id ? 'Saving...' : 'Save'}
                      </button>
                    </div>
                  </div>
                ) : (
                  <ReactMarkdown
                    remarkPlugins={[remarkGfm]}
                    components={{
                      p: ({ node: _node, ...props }) => <p className="mb-3 last:mb-0 leading-relaxed" {...props} />,
                      ul: ({ node: _node, ...props }) => <ul className="list-disc pl-5 mb-3 space-y-1" {...props} />,
                      code: ({ node: _node, ...props }) => (
                        <code className="bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded px-1 py-0.5 font-mono text-xs" {...props} />
                      ),
                    }}
                  >
                    {comment.body}
                  </ReactMarkdown>
                )}
              </div>
            </article>
          );
        })}
      </div>

      <div className="border border-slate-200 dark:border-slate-700 rounded-lg overflow-visible">
        <div className="flex border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800">
          <button
            type="button"
            onClick={() => setTab('write')}
            className={`px-4 py-2 text-sm font-medium border-r border-slate-200 dark:border-slate-700 ${
              tab === 'write' ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-white' : 'text-slate-500 dark:text-slate-400'
            }`}
          >
            Write
          </button>
          <button
            type="button"
            onClick={() => setTab('preview')}
            className={`px-4 py-2 text-sm font-medium ${
              tab === 'preview' ? 'bg-white dark:bg-slate-900 text-slate-900 dark:text-white' : 'text-slate-500 dark:text-slate-400'
            }`}
          >
            Preview
          </button>
        </div>

        {tab === 'write' && (
          <div className="flex items-center gap-1 px-2 py-2 border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 overflow-x-auto">
            <button
              type="button"
              onClick={() => applyLinePrefix('### ')}
              className="h-8 w-8 rounded text-lg font-semibold text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
              title="Heading"
            >
              H
            </button>
            <button
              type="button"
              onClick={() => applySelectionTransform((s) => ({ text: `**${s || 'bold text'}**`, cursorOffset: s ? undefined : 2 }))}
              className="h-8 w-8 rounded text-lg font-bold text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
              title="Bold"
            >
              B
            </button>
            <button
              type="button"
              onClick={() => applySelectionTransform((s) => ({ text: `*${s || 'italic text'}*`, cursorOffset: s ? undefined : 1 }))}
              className="h-8 w-8 rounded text-lg italic text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
              title="Italic"
            >
              I
            </button>
            <button
              type="button"
              onClick={() => applySelectionTransform((s) => ({ text: `\`${s || 'code'}\``, cursorOffset: s ? undefined : 1 }))}
              className="h-8 w-8 rounded text-lg text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
              title="Code"
            >
              &lt;&gt;
            </button>
            <button
              type="button"
              onClick={() =>
                applySelectionTransform((s) => ({
                  text: `[${s || 'link text'}](https://example.com)`,
                  cursorOffset: s ? undefined : 1,
                }))
              }
              className="h-8 w-8 rounded text-lg text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
              title="Link"
            >
              🔗
            </button>
            <span className="mx-1 h-6 w-px bg-slate-300 dark:bg-slate-600" />
            <div className="relative" ref={moreMenuRef}>
              <button
                type="button"
                onClick={() => setMoreMenuOpen((v) => !v)}
                className="h-8 w-8 rounded text-lg text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
                title="More"
              >
                ...
              </button>
              {moreMenuOpen && (
                <div className="absolute right-0 top-9 z-[100] w-52 overflow-hidden rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 shadow-lg">
                  <button type="button" onClick={() => handleMoreAction('unordered')} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Unordered list</button>
                  <button type="button" onClick={() => handleMoreAction('numbered')} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Numbered list</button>
                  <button type="button" onClick={() => handleMoreAction('task')} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Task list</button>
                  <div className="h-px bg-slate-200 dark:bg-slate-700" />
                  <button type="button" onClick={() => handleMoreAction('mention')} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">@ Mention</button>
                  <button type="button" onClick={() => handleMoreAction('reference')} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Reference</button>
                  <button type="button" onClick={() => handleMoreAction('slash')} className="w-full px-3 py-2 text-left text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800">Slash commands</button>
                </div>
              )}
            </div>
            <button
              type="button"
              onClick={() =>
                clearEditor({
                  draft,
                  setDraft: draftEditor.setValue,
                  undo: draftEditor.undo,
                  redo: draftEditor.redo,
                  textareaRef,
                })
              }
              className="ml-auto h-8 w-8 rounded text-base text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
              title="Clear"
              aria-label="Clear editor"
            >
              ⌫
            </button>
          </div>
        )}

        {tab === 'write' ? (
          <textarea
            ref={textareaRef}
            value={draft}
            onChange={(e) => draftEditor.setValue(e.target.value)}
            onKeyDown={(e) =>
              handleEditorKeyDown(e, {
                draft,
                setDraft: draftEditor.setValue,
                undo: draftEditor.undo,
                redo: draftEditor.redo,
                textareaRef,
              })
            }
            placeholder={`Add a comment for ${targetType} ${targetId}...`}
            className="w-full min-h-[160px] p-4 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none"
          />
        ) : (
          <div className="min-h-[160px] p-4 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200">
            {draft.trim() ? (
              <ReactMarkdown remarkPlugins={[remarkGfm]}>{draft}</ReactMarkdown>
            ) : (
              <p className="text-sm text-slate-500 dark:text-slate-400">Nothing to preview.</p>
            )}
          </div>
        )}
      </div>

      <div className="mt-3 flex justify-end">
        <button
          type="button"
          onClick={handleSubmit}
          disabled={!draft.trim() || submitting || commentActionLoadingId !== null}
          className="px-4 py-2 rounded-md bg-emerald-600 text-white text-sm font-semibold hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {submitting ? 'Saving...' : 'Comment'}
        </button>
      </div>
    </section>
  );
};

export default FeedbackSection;
