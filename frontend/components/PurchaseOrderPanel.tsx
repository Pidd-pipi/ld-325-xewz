'use client';

import { Lock, MinusCircle, RefreshCw, ShieldAlert, ShoppingCart } from 'lucide-react';
import { useEffect, useState } from 'react';

import type { PurchaseOrder, PurchaseOrderItem } from '@/lib/types';
import { money } from '@/lib/utils';
import { itemBlockReason } from '@/lib/purchase';

interface PurchaseOrderPanelProps {
  order: PurchaseOrder | null;
  onUpdateQuantity: (itemId: number, quantity: number) => Promise<void>;
  onRefresh: () => Promise<void>;
}

function differenceLabel(value: number): string {
  if (value === 0) return '持平';
  return `${value > 0 ? '+' : ''}${money(value)}`;
}

function PurchaseRow({ item, onUpdateQuantity }: { item: PurchaseOrderItem; onUpdateQuantity: PurchaseOrderPanelProps['onUpdateQuantity'] }) {
  const [quantity, setQuantity] = useState(item.quantity);
  const [saving, setSaving] = useState(false);

  useEffect(() => setQuantity(item.quantity), [item.quantity]);

  const save = async () => {
    setSaving(true);
    try {
      await onUpdateQuantity(item.id, quantity);
    } catch (error) {
      window.alert(error instanceof Error ? error.message : '数量调整失败');
    } finally {
      setSaving(false);
    }
  };

  return (
    <article className={item.purchasable ? 'purchase-item' : 'purchase-item locked'}>
      <div className="purchase-item-main">
        <span>{item.supplier_name} · 起订 {item.moq} {item.unit}</span>
        <b>{item.product_name}</b>
        <small>{item.product_model}</small>
        {!item.purchasable && <p className="purchase-warning"><ShieldAlert size={14} />{itemBlockReason(item)} · 不可采购，数量调整已暂停</p>}
      </div>
      <div className="purchase-qty">
        <label htmlFor={`purchase-qty-${item.id}`}>数量</label>
        <input
          id={`purchase-qty-${item.id}`}
          type="number"
          min={item.moq}
          value={quantity}
          disabled={!item.purchasable || saving}
          onChange={(event) => setQuantity(Number(event.target.value))}
        />
        <button disabled={!item.purchasable || saving || quantity === item.quantity || quantity < item.moq} onClick={save}>
          <RefreshCw size={13} /> 保存
        </button>
      </div>
      <div className="purchase-prices">
        <span>锁价 <strong>{money(item.locked_unit_price)}</strong></span>
        <span>当前价 <strong>{money(item.current_price)}</strong></span>
        <span>差额 <b className={item.price_difference > 0 ? 'up' : item.price_difference < 0 ? 'down' : 'flat'}>{differenceLabel(item.price_difference)}</b></span>
      </div>
    </article>
  );
}

export function PurchaseOrderPanel({ order, onUpdateQuantity, onRefresh }: PurchaseOrderPanelProps) {
  const itemCount = order?.items.length || 0;
  return (
    <section className="purchase-order" id="purchase-order">
      <div className="purchase-head">
        <div>
          <p className="eyebrow">LOCKED PURCHASE ORDER</p>
          <h2>我的采购单 <em>保留第一次锁价</em></h2>
          <p>同一商家报价重复加入时只累加数量；商家停售或资质驳回后，锁价与数量仍保留并标记为不可采购。</p>
        </div>
        <button className="purchase-refresh" onClick={onRefresh}><RefreshCw size={15} />刷新当前价</button>
      </div>
      {!order || itemCount === 0 ? (
        <div className="purchase-empty"><ShoppingCart size={28} /><b>采购单还是空的</b><span>从下方商家报价中选择“资质已通过、有货且达到起订量”的方案。</span></div>
      ) : (
        <>
          <div className="purchase-items">
            {order.items.map((item) => <PurchaseRow key={item.id} item={item} onUpdateQuantity={onUpdateQuantity} />)}
          </div>
          <div className="purchase-summary">
            <span><Lock size={15} />锁价合计 <strong>{money(order.total_locked_amount)}</strong></span>
            <span>当前合计 <strong>{money(order.total_current_amount)}</strong></span>
            <span>总差额 <b className={order.total_difference > 0 ? 'up' : order.total_difference < 0 ? 'down' : 'flat'}>{differenceLabel(order.total_difference)}</b></span>
            <span>可采购条目 <strong>{order.purchasable_count}/{itemCount}</strong></span>
          </div>
        </>
      )}
      <p className="purchase-note"><MinusCircle size={14} /> 锁价仅用于本次采购决策，合同价格仍以商家最终确认为准。</p>
    </section>
  );
}
