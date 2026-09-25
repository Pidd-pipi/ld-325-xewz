'use client';

import { useState } from 'react';
import { Ban, BellRing, CheckCircle2, Info, Lock, ShoppingCart } from 'lucide-react';

import type { Offer, Product, Trend } from '@/lib/types';
import { money } from '@/lib/utils';
import { offerPurchasable, offerStockLabel, supplierStatusLabel } from '@/lib/purchase';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { PriceTrend } from './PriceTrend';

type TrendRange = '30d' | '90d' | '1y';

interface InsightPanelProps {
  product: Product | null;
  trend: Trend | null;
  range: TrendRange;
  onRange: (range: TrendRange) => void;
  onAlert: (id: number, target: number) => Promise<void>;
  onAddOffer: (offer: Offer, quantity: number) => Promise<void>;
  addingOfferId: number | null;
}

const rangeOptions: Array<{ value: TrendRange; label: string }> = [
  { value: '30d', label: '30 天' },
  { value: '90d', label: '90 天' },
  { value: '1y', label: '1 年' },
];

export function InsightPanel({ product, trend, range, onRange, onAlert, onAddOffer, addingOfferId }: InsightPanelProps) {
  const [alerted, setAlerted] = useState(false);
  const [quantities, setQuantities] = useState<Record<number, string>>({});
  if (!product) {
    return <section className="insights"><p className="eyebrow">PRICE PULSE</p><h2>选择一款材料，查看它的价格脉搏。</h2><p>趋势、供应商、最低价和预警都将在这里展开。</p></section>;
  }

  const offers = product.Offers || [];
  const low = offers.length ? Math.min(...offers.map((offer) => offer.UnitPrice)) : 0;
  const submit = async () => {
    await onAlert(product.ID, Math.floor(low * 0.95));
    setAlerted(true);
  };

  return (
    <section className="insights">
      <div className="insight-head">
        <div>
          <p className="eyebrow">PRICE PULSE · {rangeOptions.find((option) => option.value === range)?.label}</p>
          <h2>{product.Name}</h2>
          <p>{product.Brand} · {product.Model} · 市场价按每日有效报价计算</p>
        </div>
        <Button onClick={submit} className={alerted ? 'selected' : ''}>
          {alerted ? <CheckCircle2 size={16} /> : <BellRing size={16} />} {alerted ? '预警已订阅' : `低于 ${money(Math.floor(low * 0.95))} 时提醒`}
        </Button>
      </div>
      <div className="range-tabs" role="group" aria-label="价格趋势时间范围">
        {rangeOptions.map((option) => <button key={option.value} className={range === option.value ? 'active' : ''} onClick={() => onRange(option.value)}>{option.label}</button>)}
      </div>
      <div className="trend-layout">
        <PriceTrend trend={trend} />
        <dl>
          <div><dt>{rangeOptions.find((option) => option.value === range)?.label}最低</dt><dd>{money(trend?.lowest || low)}</dd></div>
          <div><dt>{rangeOptions.find((option) => option.value === range)?.label}均价</dt><dd>{money(trend?.average || low)}</dd></div>
          <div><dt>在售商家</dt><dd>{offers.length} 家</dd></div>
        </dl>
      </div>
      <div className="offer-table">
        <div className="offer-title"><b>商家报价</b><span><Info size={14} /> 仅资质通过、有货且达到起订量可锁价采购</span></div>
        <div className="offer-row offer-head"><span>供应商</span><span>资质 / 库存</span><span>交付与起订</span><span>单价</span><span>加入采购单</span></div>
        {offers.map((offer) => {
          const purchasable = offerPurchasable(offer);
          const quantity = Number(quantities[offer.ID] ?? offer.MOQ);
          const quantityValid = Number.isInteger(quantity) && quantity >= offer.MOQ;
          return (
            <div className={offer.UnitPrice === low ? 'offer-row lowest' : 'offer-row'} key={offer.ID}>
              <b>{offer.Supplier.Name}</b>
              <span className="offer-conditions">
                <Badge tone={offer.Supplier.Status === 'approved' ? 'good' : 'alert'}>{supplierStatusLabel(offer.Supplier.Status)}</Badge>
                <Badge tone={offer.StockStatus === 'in_stock' ? 'good' : 'alert'}>{offerStockLabel(offer.StockStatus)}</Badge>
              </span>
              <span>{offer.DeliveryDays} 天交货 · 起订 {offer.MOQ} {product.Unit}<br /><small>{offer.Freight}</small></span>
              <strong>{money(offer.UnitPrice)}<small> / {product.Unit}</small></strong>
              <span className="offer-add">
                <label>
                  数量
                  <input
                    aria-label={`${offer.Supplier.Name} 采购数量`}
                    type="number" min={offer.MOQ} step="1"
                    value={quantities[offer.ID] ?? String(offer.MOQ)}
                    disabled={!purchasable || addingOfferId === offer.ID}
                    onChange={(event) => setQuantities((current) => ({ ...current, [offer.ID]: event.target.value }))}
                  />
                </label>
                <Button
                  disabled={!purchasable || !quantityValid || addingOfferId === offer.ID}
                  onClick={() => onAddOffer(offer, quantity)}
                >
                  {purchasable ? <ShoppingCart size={14} /> : <Ban size={14} />}
                  {purchasable ? '锁价加入' : '不可采购'}
                </Button>
              </span>
            </div>
          );
        })}
      </div>
      <button className="export"><Lock size={15} /> 第一次加入时锁定单价；再次加入同一报价只累加数量</button>
    </section>
  );
}
