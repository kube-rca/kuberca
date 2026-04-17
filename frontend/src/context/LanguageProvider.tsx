import { ReactNode, useEffect, useState } from 'react';
import { fetchCurrentUser, updatePreferredLanguage } from '../utils/auth';
import { Language, LanguageContext, getDictionaryValue } from './LanguageContext';

const STORAGE_KEY = 'preferred_language';

export const LanguageProvider = ({ children }: { children: ReactNode }) => {
  const [language, setLanguageState] = useState<Language>(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored === 'en' ? 'en' : 'ko';
  });

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, language);
  }, [language]);

  const setLanguage = async (nextLanguage: Language) => {
    setLanguageState(nextLanguage);
  };

  const persistLanguage = async (nextLanguage: Language) => {
    setLanguageState(nextLanguage);
    try {
      await updatePreferredLanguage(nextLanguage);
    } catch {
      // Keep optimistic UI state even if persistence fails.
    }
  };

  const toggleLanguage = async () => {
    const nextLanguage = language === 'ko' ? 'en' : 'ko';
    await persistLanguage(nextLanguage);
  };

  return (
    <LanguageContext.Provider
      value={{
        language,
        setLanguage,
        toggleLanguage,
        t: (key) => getDictionaryValue(language, key),
      }}
    >
      {children}
    </LanguageContext.Provider>
  );
};

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
