'use client';

import { Ban, ClipboardList, Lock, RefreshCw } from 'lucide-react';

import type { PurchaseOrder } from '@/lib/types';
import { money } from '@/lib/utils';
import { blockedReasonLabel, signedMoney } from '@/lib/purchase';
import { Badge } from '@/components/ui/badge';
import { QuantityControl } from '@/components/QuantityControl';

interface PurchaseOrderPanelProps {
  order: PurchaseOrder;
  updatingItemId: number | null;
  onQuantityChange: (itemId: number, quantity: number) => void;
}

export function PurchaseOrderPanel({ order, updatingItemId, onQuantityChange }: PurchaseOrderPanelProps) {
  if (!order.items.length) {
    return (
      <section className="purchase-order empty" id="purchase-order">
        <div className="purchase-heading">
          <p className="eyebrow">PURCHASE ORDER · LOCKED PRICE</p>
          <h2>采购单还是空的。</h2>
          <p>从商家报价中选择方案；首次加入时的单价会被锁定，重复加入只累加数量。</p>
        </div>
        <ClipboardList size={54} aria-hidden="true" />
      </section>
    );
  }

  return (
    <section className="purchase-order" id="purchase-order">
      <div className="purchase-heading">
        <div>
          <p className="eyebrow">PURCHASE ORDER · LOCKED PRICE</p>
          <h2>我的采购单</h2>
        </div>
        <div className="purchase-summary" aria-live="polite">
          <span><Lock size={14} /> 锁价合计 <strong>{money(order.locked_total)}</strong></span>
          <span>当前合计 <strong>{money(order.current_total)}</strong></span>
          <span>差额 <strong className={order.total_difference > 0 ? 'up' : order.total_difference < 0 ? 'down' : ''}>{signedMoney(order.total_difference)}</strong></span>
          <Badge tone={order.unavailable_count ? 'alert' : 'good'}>{order.purchasable_count} 项可采购 / {order.unavailable_count} 项暂停</Badge>
        </div>
      </div>

      <div className="purchase-list" role="table" aria-label="锁价采购单">
        <div className="purchase-row purchase-head" role="row">
          <span>报价方案</span><span>数量</span><span>锁价</span><span>当前价</span><span>差额</span><span>状态</span>
        </div>
        {order.items.map((item) => (
          <div className={`purchase-row ${item.purchasable ? '' : 'disabled'}`} role="row" key={item.id}>
            <div className="purchase-material">
              <b>{item.product_name}</b>
              <span>{item.brand} · {item.model} · {item.supplier_name}</span>
              <small>起订 {item.moq} {item.unit} · {item.delivery_days} 天交货 · {item.freight}</small>
            </div>
            <QuantityControl
              itemId={item.id} quantity={item.quantity} min={item.moq} unit={item.unit}
              disabled={!item.purchasable} busy={updatingItemId === item.id} onCommit={onQuantityChange}
            />
            <div className="locked-price"><strong>{money(item.locked_unit_price)}</strong><span>{money(item.locked_amount)}</span></div>
            <div className="current-price"><strong>{money(item.current_unit_price)}</strong><span>{money(item.current_amount)}</span></div>
            <div className={`difference ${item.amount_difference > 0 ? 'up' : item.amount_difference < 0 ? 'down' : ''}`}>
              <strong>{signedMoney(item.price_difference)}</strong>
              <span>{signedMoney(item.amount_difference)}</span>
            </div>
            <div className="purchase-status">
              {item.purchasable ? <Badge tone="good">可采购</Badge> : <Badge tone="alert"><Ban size={11} /> 不可采购</Badge>}
              {!item.purchasable && <small>{blockedReasonLabel(item.blocked_reason)}</small>}
              {updatingItemId === item.id && <RefreshCw size={13} className="spin" aria-label="正在更新" />}
            </div>
          </div>
        ))}
      </div>
      <p className="purchase-note">锁价和数量会保留；库存停售、缺货或供应商资质被驳回后，仅标记为不可采购并暂停数量调整。</p>
    </section>
  );
}
