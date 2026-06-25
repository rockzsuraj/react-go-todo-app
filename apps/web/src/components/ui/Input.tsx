import React, { useId } from 'react';
import './Input.css';

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className = '', label, error, leftIcon, rightIcon, id, ...props }, ref) => {
    const generatedId = useId();
    const inputId = id || generatedId;
    const errorId = `${inputId}-error`;

    return (
      <div className={`ui-input-wrapper ${error ? 'ui-input-wrapper--error' : ''}`}>
        {label && (
          <label htmlFor={inputId} className="ui-input-label">
            {label}
          </label>
        )}
        <div className="ui-input-container">
          {leftIcon && <span className="ui-input-icon ui-input-icon--left">{leftIcon}</span>}
          <input
            ref={ref}
            id={inputId}
            className={`ui-input ${leftIcon ? 'ui-input--has-left' : ''} ${
              rightIcon ? 'ui-input--has-right' : ''
            } ${className}`}
            aria-invalid={!!error}
            aria-describedby={error ? errorId : undefined}
            {...props}
          />
          {rightIcon && <span className="ui-input-icon ui-input-icon--right">{rightIcon}</span>}
        </div>
        {error && (
          <span id={errorId} className="ui-input-error" role="alert">
            <i className="bi bi-exclamation-circle" aria-hidden="true" />
            {error}
          </span>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';
