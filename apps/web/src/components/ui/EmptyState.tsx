import type React from 'react';
import './EmptyState.css';

export interface EmptyStateProps {
  title: string;
  description: string;
  icon?: string; // Bootstrap icon name, e.g. "bi-collection"
  action?: React.ReactNode;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  description,
  icon = 'bi-folder2-open',
  action,
}) => {
  return (
    <section className="ui-empty-state" aria-label="No content available">
      <div className="ui-empty-state__icon-wrapper">
        <i className={`bi ${icon} ui-empty-state__icon`} aria-hidden="true" />
      </div>
      <h3 className="ui-empty-state__title">{title}</h3>
      <p className="ui-empty-state__description">{description}</p>
      {action && <div className="ui-empty-state__action">{action}</div>}
    </section>
  );
};
