'use client';

import { useEffect, useState } from 'react';

interface QuantityControlProps {
  itemId: number;
  quantity: number;
  min: number;
  unit: string;
  disabled: boolean;
  busy: boolean;
  onCommit: (itemId: number, quantity: number) => void;
}

export function QuantityControl({ itemId, quantity, min, unit, disabled, busy, onCommit }: QuantityControlProps) {
  const [draft, setDraft] = useState(String(quantity));

  useEffect(() => setDraft(String(quantity)), [quantity]);

  const commit = () => {
    const next = Number(draft);
    if (Number.isInteger(next) && next >= min && next !== quantity) {
      onCommit(itemId, next);
      return;
    }
    setDraft(String(quantity));
  };

  return (
    <label className="quantity-control">
      <span className="sr-only">采购数量</span>
      <input
        type="number" min={min} step="1" value={draft}
        disabled={disabled || busy}
        onChange={(event) => setDraft(event.target.value)}
        onBlur={commit}
        onKeyDown={(event) => { if (event.key === 'Enter') event.currentTarget.blur(); }}
      />
      <em>{unit}</em>
    </label>
  );
}
