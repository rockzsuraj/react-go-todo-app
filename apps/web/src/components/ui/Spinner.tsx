import type React from 'react';
import './Spinner.css';

export interface SpinnerProps extends React.HTMLAttributes<HTMLOutputElement> {
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
    <output
      className={`ui-spinner ui-spinner--${size} ui-spinner--${variant} ${className}`}
      {...props}
    >
      <span className="visually-hidden">Loading...</span>
    </output>
  );
};
