import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import RCATable from '../components/RCATable';
import { RCAItem } from '../types';

const mockItem: RCAItem = {
  incident_id: 'inc-001',
  title: 'Test Incident',
  fired_at: '2024-01-15T10:00:00.000Z',
  resolved_at: null,
  severity: 'warning',
};

const resolvedItem: RCAItem = {
  ...mockItem,
  incident_id: 'inc-002',
  title: 'Resolved Incident',
  resolved_at: '2024-01-15T11:30:00.000Z',
  severity: 'critical',
};

describe('RCATable', () => {
  describe('empty state', () => {
    it('renders "No data available" when rcas is empty', () => {
      render(<RCATable rcas={[]} onTitleClick={vi.fn()} />);
      expect(screen.getByText('No data available')).toBeInTheDocument();
    });

    it('renders empty-state hint text', () => {
      render(<RCATable rcas={[]} onTitleClick={vi.fn()} />);
      expect(
        screen.getByText('No results match the search criteria'),
      ).toBeInTheDocument();
    });

    it('does not render table headers when empty', () => {
      render(<RCATable rcas={[]} onTitleClick={vi.fn()} />);
      expect(screen.queryByText('Title')).not.toBeInTheDocument();
    });
  });

  describe('populated state', () => {
    it('renders table headers when data is present', () => {
      render(<RCATable rcas={[mockItem]} onTitleClick={vi.fn()} />);
      expect(screen.getByText('ID')).toBeInTheDocument();
      expect(screen.getByText('Title')).toBeInTheDocument();
      expect(screen.getByText('Severity')).toBeInTheDocument();
      expect(screen.getByText('Status')).toBeInTheDocument();
    });

    it('renders incident ID in a row', () => {
      render(<RCATable rcas={[mockItem]} onTitleClick={vi.fn()} />);
      expect(screen.getByText('inc-001')).toBeInTheDocument();
    });

    it('renders multiple rows', () => {
      render(
        <RCATable rcas={[mockItem, resolvedItem]} onTitleClick={vi.fn()} />,
      );
      expect(screen.getByText('inc-001')).toBeInTheDocument();
      expect(screen.getByText('inc-002')).toBeInTheDocument();
    });

    it('formats fired_at date without T separator', () => {
      render(<RCATable rcas={[mockItem]} onTitleClick={vi.fn()} />);
      // formatDate replaces 'T' with ' ' and drops milliseconds
      expect(screen.getByText(/2024-01-15 10:00:00/)).toBeInTheDocument();
    });

    it('calls onTitleClick with correct incident_id when row is clicked', async () => {
      const onTitleClick = vi.fn();
      render(<RCATable rcas={[mockItem]} onTitleClick={onTitleClick} />);
      const row = screen.getByText('inc-001').closest('tr')!;
      await userEvent.click(row);
      expect(onTitleClick).toHaveBeenCalledWith('inc-001');
    });
  });

  describe('null-safe rendering', () => {
    it('shows "-" for null fired_at', () => {
      const item: RCAItem = { ...mockItem, fired_at: '' as string };
      render(<RCATable rcas={[item]} onTitleClick={vi.fn()} />);
      // formatDate returns '-' for falsy strings
      expect(screen.getAllByText('-').length).toBeGreaterThan(0);
    });
  });
});
