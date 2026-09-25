'use client';

import { useState } from 'react';
import { BellRing, CheckCircle2, Download, Info, Lock, Plus } from 'lucide-react';

import type { Product, Trend } from '@/lib/types';
import { money } from '@/lib/utils';
import { isOfferPurchasable, offerBlockReason } from '@/lib/purchase';
import { Button } from '@/components/ui/button';
import { PriceTrend } from './PriceTrend';

type TrendRange = '30d' | '90d' | '1y';

interface InsightPanelProps {
  product: Product | null;
  trend: Trend | null;
  range: TrendRange;
  onRange: (range: TrendRange) => void;
  onAlert: (id: number, target: number) => Promise<void>;
  onAddOffer: (offerId: number, quantity: number) => Promise<void>;
}

const rangeOptions: Array<{ value: TrendRange; label: string }> = [
  { value: '30d', label: '30 天' },
  { value: '90d', label: '90 天' },
  { value: '1y', label: '1 年' },
];

export function InsightPanel({ product, trend, range, onRange, onAlert, onAddOffer }: InsightPanelProps) {
  const [alerted, setAlerted] = useState(false);
  const [quantities, setQuantities] = useState<Record<number, number>>({});
  const [busyOffer, setBusyOffer] = useState<number | null>(null);

  if (!product) {
    return <section className="insights"><p className="eyebrow">PRICE PULSE</p><h2>选择一款材料，查看它的价格脉搏。</h2><p>趋势、供应商、最低价和预警都将在这里展开。</p></section>;
  }

  const offers = product.Offers || [];
  const low = offers.length ? Math.min(...offers.map((offer) => offer.UnitPrice)) : 0;
  const submit = async () => {
    await onAlert(product.ID, Math.floor(low * 0.95));
    setAlerted(true);
  };
  const addOffer = async (offerId: number) => {
    setBusyOffer(offerId);
    try {
      await onAddOffer(offerId, quantities[offerId] || offers.find((offer) => offer.ID === offerId)?.MOQ || 1);
    } catch (error) {
      window.alert(error instanceof Error ? error.message : '加入采购单失败');
    } finally {
      setBusyOffer(null);
    }
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
        <div className="offer-title"><b>商家报价</b><span><Info size={14} /> 第一次加入采购单时锁定当前单价</span></div>
        {offers.map((offer) => {
          const purchasable = isOfferPurchasable(offer);
          const quantity = quantities[offer.ID] ?? offer.MOQ;
          const quantityValid = quantity >= offer.MOQ;
          return (
            <div className={offer.UnitPrice === low ? 'offer lowest' : 'offer'} key={offer.ID}>
              <b>{offer.Supplier.Name}<small>{offer.Supplier.Status === 'approved' ? '资质已通过' : offer.Supplier.Status === 'rejected' ? '资质已驳回' : '资质待审核'}</small></b>
              <span>{offer.DeliveryDays} 天交货 · 起订 {offer.MOQ} {product.Unit}</span>
              <span className={purchasable ? 'stock-good' : 'stock-bad'}>{offer.StockStatus === 'in_stock' ? '有货' : offer.StockStatus === 'discontinued' ? '停售' : '缺货'}</span>
              <strong>{money(offer.UnitPrice)}</strong>
              <div className="offer-purchase">
                <label>采购数量
                  <input
                    aria-label={`${offer.Supplier.Name} 采购数量`}
                    type="number"
                    min={offer.MOQ}
                    value={quantity}
                    disabled={!purchasable || busyOffer === offer.ID}
                    onChange={(event) => setQuantities((current) => ({ ...current, [offer.ID]: Number(event.target.value) }))}
                  />
                  {product.Unit}
                </label>
                <Button disabled={!purchasable || !quantityValid || busyOffer === offer.ID} onClick={() => addOffer(offer.ID)}>
                  <Lock size={14} />{busyOffer === offer.ID ? '加入中…' : <><Plus size={14} />加入采购单</>}
                </Button>
                {!purchasable && <em>{offerBlockReason(offer)}</em>}
                {purchasable && !quantityValid && <em>数量需达到 {offer.MOQ} {product.Unit}</em>}
              </div>
            </div>
          );
        })}
      </div>
      <button className="export"><Download size={15} />导出该材料报价单</button>
    </section>
  );
}
