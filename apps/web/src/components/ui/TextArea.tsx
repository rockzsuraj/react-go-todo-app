import React, { useId } from 'react';
import './TextArea.css';

export interface TextAreaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
}

export const TextArea = React.forwardRef<HTMLTextAreaElement, TextAreaProps>(
  ({ className = '', label, error, id, ...props }, ref) => {
    const generatedId = useId();
    const inputId = id || generatedId;
    const errorId = `${inputId}-error`;

    return (
      <div className={`ui-textarea-wrapper ${error ? 'ui-textarea-wrapper--error' : ''}`}>
        {label && (
          <label htmlFor={inputId} className="ui-textarea-label">
            {label}
          </label>
        )}
        <textarea
          ref={ref}
          id={inputId}
          className={`ui-textarea ${className}`}
          aria-invalid={!!error}
          aria-describedby={error ? errorId : undefined}
          {...props}
        />
        {error && (
          <span id={errorId} className="ui-textarea-error" role="alert">
            <i className="bi bi-exclamation-circle" aria-hidden="true" />
            {error}
          </span>
        )}
      </div>
    );
  }
);

TextArea.displayName = 'TextArea';
