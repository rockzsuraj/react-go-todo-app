import React, { useId } from 'react';
import './Select.css';

export interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  error?: string;
  leftIcon?: React.ReactNode;
}

export const Select = React.forwardRef<HTMLSelectElement, SelectProps>(
  ({ className = '', label, error, leftIcon, id, children, ...props }, ref) => {
    const generatedId = useId();
    const selectId = id || generatedId;
    const errorId = `${selectId}-error`;

    return (
      <div className={`ui-select-wrapper ${error ? 'ui-select-wrapper--error' : ''}`}>
        {label && (
          <label htmlFor={selectId} className="ui-select-label">
            {label}
          </label>
        )}
        <div className="ui-select-container">
          {leftIcon && <span className="ui-select-icon ui-select-icon--left">{leftIcon}</span>}
          <select
            ref={ref}
            id={selectId}
            className={`ui-select ${leftIcon ? 'ui-select--has-left' : ''} ${className}`}
            aria-invalid={!!error}
            aria-describedby={error ? errorId : undefined}
            {...props}
          >
            {children}
          </select>
        </div>
        {error && (
          <span id={errorId} className="ui-select-error" role="alert">
            <i className="bi bi-exclamation-circle" aria-hidden="true" />
            {error}
          </span>
        )}
      </div>
    );
  }
);

Select.displayName = 'Select';
