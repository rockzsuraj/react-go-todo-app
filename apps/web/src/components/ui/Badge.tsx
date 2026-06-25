import React from 'react';
import './Badge.css';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'success' | 'warning' | 'info' | 'neutral' | 'error';
}

export const Badge: React.FC<BadgeProps> = ({
  className = '',
  variant = 'neutral',
  children,
  ...props
}) => {
  return (
    <span className={`ui-badge ui-badge--${variant} ${className}`} {...props}>
      {children}
    </span>
  );
};
