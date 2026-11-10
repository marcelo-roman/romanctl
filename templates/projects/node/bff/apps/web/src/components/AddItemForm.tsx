import { type FormEvent, useState } from 'react';
import type { ProductSuggestion } from '../bff/orders';

interface AddItemFormProps {
  suggestions: ProductSuggestion[];
  pending: boolean;
  onAdd: (productId: string, quantity: number) => Promise<void>;
}

export function AddItemForm({ suggestions, pending, onAdd }: AddItemFormProps) {
  const [productId, setProductId] = useState(suggestions[0]?.id ?? '');
  const [quantity, setQuantity] = useState('1');

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onAdd(productId, Number(quantity));
  }

  return (
    <form className="inline-form" onSubmit={submit}>
      <label>
        Produto
        <select value={productId} onChange={(event) => setProductId(event.target.value)}>
          {suggestions.map((product) => (
            <option key={product.id} value={product.id}>
              {product.name} ({product.price})
            </option>
          ))}
        </select>
      </label>
      <label>
        Quantidade
        <input type="number" value={quantity} onChange={(event) => setQuantity(event.target.value)} />
      </label>
      <button type="submit" disabled={pending}>
        Adicionar
      </button>
    </form>
  );
}
