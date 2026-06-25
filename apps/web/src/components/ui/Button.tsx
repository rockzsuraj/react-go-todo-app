import React from 'react';
import './Button.css';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className = '',
      variant = 'primary',
      size = 'md',
      isLoading = false,
      leftIcon,
      rightIcon,
      children,
      disabled,
      type = 'button',
      ...props
    },
    ref
  ) => {
    return (
      <button
        ref={ref}
        type={type}
        disabled={disabled || isLoading}
        className={`ui-btn ui-btn--${variant} ui-btn--${size} ${isLoading ? 'ui-btn--loading' : ''} ${className}`}
        aria-busy={isLoading}
        {...props}
      >
        {isLoading && (
          <span className="ui-btn__spinner" aria-hidden="true">
            <span className="ui-btn__spinner-inner" />
          </span>
        )}
        {!isLoading && leftIcon && <span className="ui-btn__icon ui-btn__icon--left">{leftIcon}</span>}
        <span className="ui-btn__text">{children}</span>
        {!isLoading && rightIcon && <span className="ui-btn__icon ui-btn__icon--right">{rightIcon}</span>}
      </button>
    );
  }
);

Button.displayName = 'Button';
