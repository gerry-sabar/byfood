import { InputHTMLAttributes, forwardRef } from 'react';
import clsx from 'clsx';

type Props = InputHTMLAttributes<HTMLInputElement> & { invalid?: boolean };

const Input = forwardRef<HTMLInputElement, Props>(function Input({ className, invalid, ...props }, ref) {
  return (
    <input
      ref={ref}
      className={clsx('input-base', invalid ? 'input-error' : 'input-default', className)}
      {...props}
    />
  );
});

export default Input;