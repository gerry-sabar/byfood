import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';
import BooksPageTemplate from '@/templates/BooksPageTemplate';
import { api } from '@/lib/api';
import { Book } from '@/types/book';

export default function BookDetail() {
  const router = useRouter();
  const { id } = router.query as { id?: string };
  const [book, setBook] = useState<Book | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!id) return;
    (async () => {
      setLoading(true);
      try {
        const data = await api.getBook(id);
        setBook(data);
        setError(null);
      } catch (e: any) {
        setBook(null);
        setError(e.message || 'Book not found');
      } finally {
        setLoading(false);
      }
    })();
  }, [id]);

  return (
    <BooksPageTemplate>
      <button className="btn btn-ghost mb-4" onClick={() => router.push('/')}>
        ← Back
      </button>

      {loading && <div className="text-slate-600">Loading…</div>}

      {!loading && error && (
        <div className="card p-8 text-center space-y-4">
          <h2 className="text-2xl font-bold text-slate-800">Book not found</h2>
          <p className="text-slate-600">
            We couldn’t find a book with ID <span className="font-mono">{id}</span>.
          </p>
          <button
            className="btn btn-primary"
            onClick={() => router.push('/')}
          >
            Back to Library
          </button>
        </div>
      )}

      {!loading && book && (
        <div className="card p-6 space-y-2">
          <h2 className="text-2xl font-bold">{book.title}</h2>
          <p className="text-slate-700">By {book.author}</p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-4">
            <div className="p-3 rounded-xl bg-slate-50 border">
              ISBN: <span className="font-mono">{book.isbn}</span>
            </div>
            <div className="p-3 rounded-xl bg-slate-50 border">
              Price:{' '}
              {typeof book.price === 'number'
                ? book.price.toFixed(2)
                : book.price}
            </div>
            {book.created_at && (
              <div className="p-3 rounded-xl bg-slate-50 border">
                Created: {new Date(book.created_at).toLocaleString()}
              </div>
            )}
            <div className="p-3 rounded-xl bg-slate-50 border">
              Publication Year:{' '}
              {typeof book.publication_year === 'number'
                ? book.publication_year.toFixed(2)
                : book.publication_year}
            </div>
          </div>
        </div>
      )}
    </BooksPageTemplate>
  );
}
