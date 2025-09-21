import { useState, useMemo } from 'react';
import { Book } from '@/types/book';
import Modal from '@/components/atoms/Modal';
import Button from '@/components/atoms/Button';

export type BookTableProps = {
  books?: Book[] | null;        // ← tolerate undefined/null
  onEdit: (book: Book) => void;
  onDelete: (book: Book) => void;
};

export default function BookTable({ books, onEdit, onDelete }: BookTableProps) {
  const [deleting, setDeleting] = useState<Book | null>(null);

  // Always work with a safe array
  const rows = useMemo(() => (Array.isArray(books) ? books : []), [books]);

  return (
    <div className="card overflow-hidden">
      <table className="w-full text-sm">
        <thead className="bg-slate-50 text-slate-600">
          <tr>
            <th className="text-left px-4 py-2">Title</th>
            <th className="text-left px-4 py-2">Author</th>
            <th className="text-left px-4 py-2">ISBN</th>
            <th className="text-right px-4 py-2">Price</th>
            <th className="text-center px-4 py-2">Publication Year</th>
            <th className="text-center px-4 py-2">Actions</th>
          </tr>
        </thead>

        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td className="px-4 py-6 text-center text-slate-500" colSpan={5}>
                No books found.
              </td>
            </tr>
          ) : (
            rows.map((b) => (
              <tr key={b.id} className="border-t">
                <td className="px-4 py-2">
                  <a className="text-blue-400 underline" href={`/books/${b.id}`}>
                    {b.title}
                  </a>
                </td>
                <td className="px-4 py-2">{b.author}</td>
                <td className="px-4 py-2">{b.isbn}</td>
                <td className="px-4 py-2 text-right">
                  {typeof b.price === 'number' ? b.price.toFixed(2) : b.price ?? '—'}
                </td>
                <td className="px-4 py-2 text-center">{b.publication_year}</td>
                <td className="px-4 py-2 text-right">
                  <button className="btn btn-ghost mr-2" onClick={() => onEdit(b)}>
                    Edit
                  </button>
                  <button
                    className="btn btn-ghost text-red-600"
                    onClick={() => setDeleting(b)}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>

      {/* Delete confirmation modal */}
      <Modal open={!!deleting} onClose={() => setDeleting(null)} title="Confirm Delete">
        {deleting && (
          <div className="space-y-4">
            <p>
              Are you sure you want to delete{' '}
              <span className="font-semibold">“{deleting.title}”</span>?
            </p>
            <div className="flex justify-end gap-2">
              <button className="btn btn-ghost" onClick={() => setDeleting(null)}>
                Cancel
              </button>
              <Button
                className="bg-red-600 border-red-600 hover:opacity-90"
                onClick={() => {
                  onDelete(deleting);
                  setDeleting(null);
                }}
              >
                Delete
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
