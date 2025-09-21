import { useEffect, useState } from 'react';
import BooksPageTemplate from '@/templates/BooksPageTemplate';
import BookTable from '@/components/organisms/BookTable';
import BookFormModal from '@/components/organisms/BookFormModal';
import Modal from '@/components/atoms/Modal';
import Button from '@/components/atoms/Button';
import { api } from '@/lib/api';
import { Book } from '@/types/book';

type NormalizedApiError = {
  title: string;
  message?: string;
  fields?: Record<string, string>;
};

function normalizeError(e: unknown): NormalizedApiError {
  // api.ts throws `await res.text()` -> may be plain text or JSON string
  let raw = typeof e === 'string' ? e : (e as any)?.message || 'Unexpected error';
  let payload: any = null;
  try { payload = JSON.parse(raw); } catch { /* ignore */ }

  // Shapes we expect from the Go backend:
  // - { error: "duplicate isbn" }
  // - { fields: { isbn: "already exists" } }
  // - plain string
  if (payload && typeof payload === 'object') {
    const title =
      payload.error?.toString() ||
      'Request failed';
    const fields = typeof payload.fields === 'object' ? payload.fields : undefined;

    return {
      title,
      message: fields ? 'Please fix the highlighted fields.' : undefined,
      fields,
    };
  }

  // If the backend sent a known duplicate ISBN message, make it friendly
  if (/duplicate/i.test(raw) && /isbn/i.test(raw)) {
    return { title: 'Duplicate ISBN', message: 'A book with this ISBN already exists.' };
  }

  return { title: 'Request failed', message: raw };
}

export default function Home() {
  const [books, setBooks] = useState<Book[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Book | null>(null);

  const [success, setSuccess] = useState<string | null>(null);
  const [errDialog, setErrDialog] = useState<NormalizedApiError | null>(null);

  async function refresh() {
    try {
      setLoading(true);
      const data = await api.listBooks();
      setBooks(Array.isArray(data) ? data : []);
      setError(null);
    } catch (e: any) {
      setError(e.message || 'Failed to load');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    refresh();
  }, []);

  return (
    <BooksPageTemplate>
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Library</h2>
        <Button
          onClick={() => {
            setEditing(null);
            setOpen(true);
          }}
        >
          Add new book
        </Button>
      </div>

      {loading && <div className="text-slate-600">Loading…</div>}
      {error && <div className="text-red-600">{error}</div>}

      {!loading && !error && (
        <BookTable
          books={books}
          onEdit={(b) => {
            setEditing(b);
            setOpen(true);
          }}
          onDelete={async (b) => {
            try {
              await api.deleteBook(b.id);
              await refresh();
              setSuccess('Book deleted successfully ✅');
            } catch (e) {
              setErrDialog(normalizeError(e));
            }
          }}
        />
      )}

      <BookFormModal
        open={open}
        initial={editing || undefined}
        onClose={() => {
          setOpen(false);
          setEditing(null);
        }}
        onSubmit={async (values) => {
          try {
            if (editing) {
              await api.updateBook(editing.id, values);
              setSuccess('Book updated successfully ✅');
            } else {
              await api.createBook(values);
              setSuccess('Book created successfully ✅');
            }
            await refresh();
            setOpen(false);
            setEditing(null);
          } catch (e) {
            // 🔴 Prevent React overlay: handle and show friendly dialog
            setErrDialog(normalizeError(e));
          }
        }}
      />

      {/* ✅ Success dialog */}
      <Modal open={!!success} onClose={() => setSuccess(null)} title="Success">
        <div className="space-y-4 text-center">
          <p className="text-green-700 font-medium">{success}</p>
          <div className="flex justify-center">
            <Button onClick={() => setSuccess(null)}>OK</Button>
          </div>
        </div>
      </Modal>

      {/* 🔴 Error dialog */}
      <Modal
        open={!!errDialog}
        onClose={() => setErrDialog(null)}
        title="There was a problem"
      >
        {errDialog && (
          <div className="space-y-4">
            <p className="text-red-700 font-medium">{errDialog.title}</p>
            {errDialog.message && (
              <p className="text-slate-700">{errDialog.message}</p>
            )}
            {errDialog.fields && (
              <div className="rounded-xl bg-red-50 border border-red-200 p-3">
                <h4 className="font-semibold text-red-700 mb-2">Field errors</h4>
                <ul className="list-disc list-inside text-sm text-red-700">
                  {Object.entries(errDialog.fields).map(([k, v]) => (
                    <li key={k}>
                      <span className="font-mono">{k}</span>: {String(v)}
                    </li>
                  ))}
                </ul>
              </div>
            )}
            <div className="flex justify-end">
              <Button className="bg-red-600 border-red-600" onClick={() => setErrDialog(null)}>
                Close
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </BooksPageTemplate>
  );
}
