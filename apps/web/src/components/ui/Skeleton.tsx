import type React from 'react';
import './Skeleton.css';

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  width?: string | number;
  height?: string | number;
  variant?: 'text' | 'rect' | 'circle';
}

export const Skeleton: React.FC<SkeletonProps> = ({
  className = '',
  width,
  height,
  variant = 'rect',
  style,
  ...props
}) => {
  const customStyles: React.CSSProperties = {
    width,
    height,
    ...style,
  };

  return (
    <div
      className={`ui-skeleton ui-skeleton--${variant} ${className}`}
      style={customStyles}
      {...props}
    />
  );
};
