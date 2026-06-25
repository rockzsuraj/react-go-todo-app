import React from 'react';
import './Spinner.css';

export interface SpinnerProps extends React.HTMLAttributes<HTMLSpanElement> {
  size?: 'sm' | 'md' | 'lg';
  variant?: 'primary' | 'neutral' | 'white';
}

export const Spinner: React.FC<SpinnerProps> = ({
  className = '',
  size = 'md',
  variant = 'primary',
  ...props
}) => {
  return (
    <span
      className={`ui-spinner ui-spinner--${size} ui-spinner--${variant} ${className}`}
      role="status"
      {...props}
    >
      <span className="visually-hidden">Loading...</span>
    </span>
  );
};
