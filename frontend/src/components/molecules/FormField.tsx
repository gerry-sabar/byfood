import { ReactNode } from 'react';
import clsx from 'clsx';

type Props = {
  label: string;
  htmlFor?: string;
  help?: string;
  error?: string;
  children: ReactNode;
};

export default function FormField({ label, htmlFor, help, error, children }: Props) {
  return (
    <div>
      <label className="label" htmlFor={htmlFor}>{label}</label>
      {children}
      <p className={clsx('help', error ? 'help-error' : undefined)}>
        {error ? error : help}
      </p>
    </div>
  );
}