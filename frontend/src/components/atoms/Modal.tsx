import { ReactNode, useEffect } from 'react';

type Props = { open: boolean; onClose: () => void; title?: string; children: ReactNode };

export default function Modal({ open, onClose, title, children }: Props) {
  useEffect(() => {
    function onEsc(e: KeyboardEvent) { if (e.key === 'Escape') onClose(); }
    if (open) document.addEventListener('keydown', onEsc);
    return () => document.removeEventListener('keydown', onEsc);
  }, [open, onClose]);

  if (!open) return null;
  return (
    <div className="modal-panel">
      <div className="modal-backdrop" onClick={onClose} />
      <div className="card w-full max-w-lg p-4 relative z-10">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">{title}</h3>
          <button aria-label="Close" className="btn btn-ghost px-2 py-1" onClick={onClose}>✕</button>
        </div>
        <div className="mt-3">{children}</div>
      </div>
    </div>
  );
}