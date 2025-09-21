import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Modal from '@/components/atoms/Modal';
import Input from '@/components/atoms/Input';
import Button from '@/components/atoms/Button';
import FormField from '@/components/molecules/FormField';
import { bookSchema, BookFormValues } from '@/utils/validate';

export type BookFormModalProps = {
  open: boolean;
  initial?: Partial<BookFormValues> & { id?: number };
  onClose: () => void;
  onSubmit: (values: BookFormValues) => Promise<void> | void;
};

const DEFAULTS: BookFormValues = { title: '', author: '', isbn: '', publication_year: new Date().getFullYear(), price: 0 };

export default function BookFormModal({ open, initial, onClose, onSubmit }: BookFormModalProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<BookFormValues>({
    resolver: zodResolver(bookSchema),
    defaultValues: DEFAULTS,
    mode: 'onChange',
    reValidateMode: 'onChange',
  });

  useEffect(() => {
    if (open) {
      const values = { ...DEFAULTS, ...(initial ?? {}) };
      reset(values, {
        keepErrors: false,
        keepDirty: false,
        keepTouched: false,
        keepIsSubmitted: false,
        keepSubmitCount: false,
      });
    }
  }, [open, initial, reset]);

  return (
    <Modal
      open={open}
      onClose={() => {
        reset(DEFAULTS);
        onClose();
      }}
      title={initial?.id ? 'Edit Book' : 'Add Book'}
    >
      <form
        className="space-y-3"
        onSubmit={handleSubmit(async (values) => {
          await onSubmit(values);
          reset(DEFAULTS);
          onClose();
        })}
      >
        <FormField label="Title" error={errors.title?.message}>
          <Input {...register('title')} placeholder="Clean Code" invalid={!!errors.title} />
        </FormField>

        <FormField label="Author" error={errors.author?.message}>
          <Input {...register('author')} placeholder="Robert C. Martin" invalid={!!errors.author} />
        </FormField>

        <FormField label="ISBN" error={errors.isbn?.message}>
          <Input {...register('isbn')} placeholder="9780132350884" invalid={!!errors.isbn} />
        </FormField>

        <FormField label="Publication Year" error={errors.publication_year?.message}>
          <Input
            type="number"
            inputMode="numeric"
            pattern="\d*"
            {...register('publication_year', {
              valueAsNumber: true,
              setValueAs: (v) => (v === '' || v === null ? undefined : Number(v)),
            })}
            placeholder="2008"
            invalid={!!errors.publication_year}
          />
        </FormField>

        <FormField label="Price" error={errors.price?.message}>
          <Input
            type="number"
            step="0.01"
            {...register('price', { valueAsNumber: true })}
            placeholder="0.00"
            invalid={!!errors.price}
          />
        </FormField>

        <div className="flex justify-end gap-2 pt-2">
          <button
            type="button"
            className="btn btn-ghost"
            onClick={() => {
              reset(DEFAULTS);
              onClose();
            }}
          >
            Cancel
          </button>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Saving…' : 'Save'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}
