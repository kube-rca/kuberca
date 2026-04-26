import { ReactNode, useEffect, useState } from 'react';
import { updatePreferredLanguage } from '../utils/auth';
import { Language, LanguageContext, getDictionaryValue } from './LanguageContext';
import { languageStorageKey } from './languageStorage';

export const LanguageProvider = ({ children }: { children: ReactNode }) => {
  const [language, setLanguageState] = useState<Language>(() => {
    const stored = localStorage.getItem(languageStorageKey);
    return stored === 'en' ? 'en' : 'ko';
  });

  useEffect(() => {
    localStorage.setItem(languageStorageKey, language);
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
