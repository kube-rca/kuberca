import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import AuthPanel from '../components/AuthPanel';
import { ThemeProvider } from '../context/ThemeProvider';
import { LanguageProvider } from '../context/LanguageProvider';

// AuthPanel requires both ThemeContext and LanguageContext.
const Providers = ({ children }: { children: React.ReactNode }) => (
  <ThemeProvider>
    <LanguageProvider>{children}</LanguageProvider>
  </ThemeProvider>
);

const defaultProps = {
  allowSignup: false,
  oidcEnabled: false,
  oidcLoginUrl: '',
  oidcProvider: 'google',
  onAuthenticated: vi.fn(),
};

describe('AuthPanel', () => {
  describe('branding', () => {
    it('renders the Kube-RCA application title', () => {
      render(
        <Providers>
          <AuthPanel {...defaultProps} />
        </Providers>,
      );
      expect(screen.getByText('Kube-RCA')).toBeInTheDocument();
    });

    it('renders the Kube-RCA logo image', () => {
      render(
        <Providers>
          <AuthPanel {...defaultProps} />
        </Providers>,
      );
      expect(screen.getByAltText('Kube-RCA')).toBeInTheDocument();
    });
  });

  describe('language selector', () => {
    it('renders the Language label', () => {
      render(
        <Providers>
          <AuthPanel {...defaultProps} />
        </Providers>,
      );
      expect(screen.getByText('Language')).toBeInTheDocument();
    });

    it('shows both ko and en language options', () => {
      render(
        <Providers>
          <AuthPanel {...defaultProps} />
        </Providers>,
      );
      expect(screen.getByRole('option', { name: '한국어' })).toBeInTheDocument();
      expect(screen.getByRole('option', { name: 'English' })).toBeInTheDocument();
    });
  });

  describe('OIDC disabled (default credential login)', () => {
    it('does not render OIDC sign-in button when oidcEnabled is false', () => {
      render(
        <Providers>
          <AuthPanel {...defaultProps} oidcEnabled={false} />
        </Providers>,
      );
      expect(screen.queryByText('Sign in with Google')).not.toBeInTheDocument();
    });

    it('renders login form fields', () => {
      render(
        <Providers>
          <AuthPanel {...defaultProps} />
        </Providers>,
      );
      // ID and password inputs must be present for credential login.
      expect(screen.getAllByRole('textbox').length).toBeGreaterThanOrEqual(1);
    });
  });

  describe('OIDC enabled', () => {
    it('renders the OIDC sign-in button when oidcEnabled is true', () => {
      render(
        <Providers>
          <AuthPanel
            {...defaultProps}
            oidcEnabled={true}
            oidcLoginUrl="/api/v1/auth/oidc/login"
            oidcProvider="google"
          />
        </Providers>,
      );
      expect(screen.getByText('Sign in with Google')).toBeInTheDocument();
    });

    it('renders generic SSO label for unknown provider', () => {
      render(
        <Providers>
          <AuthPanel
            {...defaultProps}
            oidcEnabled={true}
            oidcLoginUrl="/api/v1/auth/oidc/login"
            oidcProvider="unknown-provider"
          />
        </Providers>,
      );
      expect(screen.getByText('Sign in with SSO')).toBeInTheDocument();
    });

    it('renders Azure provider label', () => {
      render(
        <Providers>
          <AuthPanel
            {...defaultProps}
            oidcEnabled={true}
            oidcLoginUrl="/api/v1/auth/oidc/login"
            oidcProvider="azure"
          />
        </Providers>,
      );
      expect(screen.getByText('Sign in with Microsoft')).toBeInTheDocument();
    });
  });

  describe('signup toggle', () => {
    it('does not show signup link when allowSignup is false', () => {
      render(
        <Providers>
          <AuthPanel {...defaultProps} allowSignup={false} />
        </Providers>,
      );
      expect(screen.queryByText(/계정이 없나요/)).not.toBeInTheDocument();
    });

    it('shows signup link when allowSignup is true', () => {
      render(
        <Providers>
          <AuthPanel {...defaultProps} allowSignup={true} />
        </Providers>,
      );
      expect(screen.getByText(/계정이 없나요/)).toBeInTheDocument();
    });
  });

  describe('theme toggle', () => {
    it('renders toggle dark mode button', () => {
      render(
        <Providers>
          <AuthPanel {...defaultProps} />
        </Providers>,
      );
      expect(
        screen.getByRole('button', { name: /toggle dark mode/i }),
      ).toBeInTheDocument();
    });
  });
});
