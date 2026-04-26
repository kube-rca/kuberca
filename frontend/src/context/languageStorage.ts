import { fetchCurrentUser } from '../utils/auth';
import { Language } from './LanguageContext';

const STORAGE_KEY = 'preferred_language';

export const languageStorageKey = STORAGE_KEY;

export const loadStoredLanguage = (): Language =>
  localStorage.getItem(STORAGE_KEY) === 'en' ? 'en' : 'ko';

export const syncLanguageFromServer = async (): Promise<Language> => {
  try {
    const user = await fetchCurrentUser();
    return user.preferredLanguage === 'en' ? 'en' : 'ko';
  } catch {
    return loadStoredLanguage();
  }
};
