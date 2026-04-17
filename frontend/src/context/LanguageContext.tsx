import { createContext, useContext } from 'react';

export type Language = 'ko' | 'en';

export const languageLabels: Record<Language, string> = {
  ko: '한국어',
  en: 'English',
};

type Dictionary = Record<string, string>;

const dictionaries: Record<Language, Dictionary> = {
  ko: {
    loginDescription: 'ID와 비밀번호로 로그인하세요.',
    signupDescription: '새 계정을 생성하세요.',
    id: '아이디',
    username: '사용자 이름',
    password: '비밀번호',
    processing: '처리 중...',
    login: '로그인',
    signUp: '회원가입',
    noAccount: '계정이 없나요?',
    alreadyAccount: '이미 계정이 있나요?',
    logout: '로그아웃',
    incidents: '인시던트',
    alerts: '알림',
    archived: '보관됨',
    analysis: '분석',
    settings: '설정',
    agentChat: '에이전트 채팅',
    chatIntro: '질문을 입력하면 Incident/Alert 컨텍스트를 함께 분석해서 답변합니다.',
    page: '페이지',
    context: '컨텍스트',
    collapse: '접기',
    expand: '확장',
    close: '닫기',
    incidentOptional: 'incident_id (선택)',
    alertOptional: 'alert_id (선택)',
    contextHint: '상세 페이지에서는 ID가 자동 입력되며, 대시보드에서는 정확도를 높이기 위해 직접 입력할 수 있습니다.',
    generating: '답변 생성 중...',
    closeChat: '채팅 닫기',
    aiChat: 'AI 채팅',
  },
  en: {
    loginDescription: 'Log in with your ID and password.',
    signupDescription: 'Create a new account.',
    id: 'ID',
    username: 'Username',
    password: 'Password',
    processing: 'Processing...',
    login: 'Login',
    signUp: 'Sign up',
    noAccount: "Don't have an account?",
    alreadyAccount: 'Already have an account?',
    logout: 'Logout',
    incidents: 'Incidents',
    alerts: 'Alerts',
    archived: 'Archived',
    analysis: 'Analysis',
    settings: 'Settings',
    agentChat: 'Agent Chat',
    chatIntro: 'If you enter a question, it will analyze the Incident/Alert context together and answer.',
    page: 'Page',
    context: 'Context',
    collapse: 'Collapse',
    expand: 'Expand',
    close: 'Close',
    incidentOptional: 'incident_id (optional)',
    alertOptional: 'alert_id (optional)',
    contextHint: 'The ID is automatically filled in on the details page, and you can manually enter it on the dashboard to increase accuracy.',
    generating: 'Generating answer...',
    closeChat: 'Close Chat',
    aiChat: 'AI Chat',
  },
};

interface LanguageContextType {
  language: Language;
  setLanguage: (language: Language) => Promise<void> | void;
  toggleLanguage: () => Promise<void> | void;
  t: (key: keyof typeof dictionaries.en) => string;
}

export const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

export const useLanguage = () => {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
};

export const getDictionaryValue = (language: Language, key: keyof typeof dictionaries.en) =>
  dictionaries[language][key] || dictionaries.en[key];
