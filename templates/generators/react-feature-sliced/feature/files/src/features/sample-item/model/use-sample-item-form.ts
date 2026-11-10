import { useMutation } from '@tanstack/react-query';
import { useState, type FormEvent } from 'react';
import { errorMessage } from '@/shared/api';
import { submitSampleItem } from '../api/submit-sample-item';
import { emptySampleItem, toSampleItemInput, validateSampleItem, type SampleItemValues } from './sample-item-form';

export function useSampleItemForm() {
  const [values, setValues] = useState<SampleItemValues>(emptySampleItem);
  const [invalid, setInvalid] = useState<string>();
  const mutation = useMutation({
    mutationFn: submitSampleItem,
    onSuccess: () => setValues(emptySampleItem),
  });

  const change = (field: keyof SampleItemValues) => (value: string) => {
    setValues((current) => ({ ...current, [field]: value }));
  };

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const violation = validateSampleItem(values);
    setInvalid(violation);
    if (violation) {
      return;
    }
    mutation.mutate(toSampleItemInput(values));
  };

  return {
    values,
    change,
    submit,
    isPending: mutation.isPending,
    succeeded: mutation.isSuccess,
    error: invalid ?? (mutation.isError ? errorMessage(mutation.error) : undefined),
  };
}
