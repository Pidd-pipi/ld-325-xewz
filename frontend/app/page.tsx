'use client';

import { useEffect, useMemo, useState } from 'react';
import { ArrowDownRight, LoaderCircle, ShieldCheck } from 'lucide-react';

import { api } from '@/lib/api';
import type { Offer, Product, PurchaseOrder, Trend } from '@/lib/types';
import { offerPurchasable } from '@/lib/purchase';
import { AppHeader } from '@/components/AppHeader';
import { CategoryRail } from '@/components/CategoryRail';
import { ProductCatalog } from '@/components/ProductCatalog';
import { ComparisonTray } from '@/components/ComparisonTray';
import { InsightPanel } from '@/components/InsightPanel';
import { BudgetCalculator } from '@/components/BudgetCalculator';
import { PurchaseOrderPanel } from '@/components/PurchaseOrderPanel';

const emptyOrder: PurchaseOrder = {
  id: 0, items: [], locked_total: 0, current_total: 0,
  total_difference: 0, purchasable_count: 0, unavailable_count: 0,
};

export default function Home() {
  const [products, setProducts] = useState<Product[]>([]);
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState('全部');
  const [compareIDs, setCompareIDs] = useState<number[]>([]);
  const [selected, setSelected] = useState<Product | null>(null);
  const [trend, setTrend] = useState<Trend | null>(null);
  const [trendRange, setTrendRange] = useState<'30d' | '90d' | '1y'>('30d');
  const [sort, setSort] = useState<'price' | 'sales' | 'rating'>('rating');
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [order, setOrder] = useState<PurchaseOrder>(emptyOrder);
  const [addingOfferId, setAddingOfferId] = useState<number | null>(null);
  const [updatingItemId, setUpdatingItemId] = useState<number | null>(null);

  useEffect(() => {
    Promise.all([api.listProducts(), api.getPurchaseOrder()])
      .then(([catalog, purchaseOrder]) => {
        setProducts(catalog.items);
        setSelected(catalog.items[0] || null);
        setOrder(purchaseOrder);
      })
      .catch(() => setMessage('报价数据暂时不可用，请确认后端服务已启动。'))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (selected) api.trend(selected.ID, trendRange).then(setTrend).catch(() => setTrend(null));
  }, [selected, trendRange]);

  const visible = useMemo(() => products
    .filter((item) => (category === '全部' || item.Category?.Name === category)
      && `${item.Name}${item.Brand}${item.Model}`.toLowerCase().includes(query.toLowerCase()))
    .sort((a, b) => sort === 'price'
      ? Math.min(...a.Offers.map((offer) => offer.UnitPrice)) - Math.min(...b.Offers.map((offer) => offer.UnitPrice))
      : sort === 'sales' ? b.SalesCount - a.SalesCount : b.Rating - a.Rating),
    [products, category, query, sort]);

  const notify = (text: string) => {
    setMessage(text);
    window.setTimeout(() => setMessage(''), 3200);
  };

  const compare = (id: number) => setCompareIDs((current) => (
    current.includes(id) ? current.filter((value) => value !== id) : current.length < 4 ? [...current, id] : current
  ));

  const addOfferToOrder = async (offer: Offer, quantity: number) => {
    if (!offerPurchasable(offer) || quantity < offer.MOQ) {
      notify('仅可选择资质已通过、有货且达到起订量的报价');
      return;
    }
    setAddingOfferId(offer.ID);
    try {
      setOrder(await api.addPurchaseItem(offer.ID, quantity));
      notify(`已按首次锁价加入采购单，数量 ${quantity}，重复选择会累加`);
    } catch (error) {
      notify(error instanceof Error ? error.message : '加入采购单失败');
    } finally {
      setAddingOfferId(null);
    }
  };

  const changeOrderQuantity = async (itemId: number, quantity: number) => {
    setUpdatingItemId(itemId);
    try {
      setOrder(await api.updatePurchaseItem(itemId, quantity));
    } catch (error) {
      notify(error instanceof Error ? error.message : '数量调整失败');
      setOrder(await api.getPurchaseOrder());
    } finally {
      setUpdatingItemId(null);
    }
  };

  return (
    <main id="top">
      <AppHeader query={query} onQuery={setQuery} orderCount={order.items.length} />
      <section className="hero">
        <div>
          <p className="eyebrow">MATERIAL MARKET INTELLIGENCE / SINCE 2026</p>
          <h1>不是找最低价。<br /><em>是买到恰好的那一笔。</em></h1>
          <p className="hero-copy">把品牌、规格、交期与多商家报价放到同一张桌子上。锁定第一次看到的单价，后续涨跌都有账可查。</p>
          <div className="hero-note"><ShieldCheck size={18} /><span>演示数据每日价格记录 · 30 天历史趋势 · 采购单锁价保留</span></div>
        </div>
        <div className="hero-number">
          <span>本期已收录</span><strong>8,624</strong><b>条有效报价 <ArrowDownRight size={18} /></b>
          <p>覆盖瓷砖、地板、涂料、卫浴等<br />装修决策中的高频材料。</p>
        </div>
      </section>
      <CategoryRail selected={category} onSelected={setCategory} />
      {loading ? (
        <div className="loading"><LoaderCircle className="spin" /> 正在汇总市场报价…</div>
      ) : (
        <>
          <ProductCatalog
            products={visible} compareIDs={compareIDs} onCompare={compare}
            onFavorite={async (id) => { await api.favorite(id); notify('已收入「本周采购」收藏夹'); }}
            onSelect={setSelected} sort={sort} onSort={setSort}
          />
          <ComparisonTray items={products.filter((item) => compareIDs.includes(item.ID))} onRemove={compare} />
          <InsightPanel
            product={selected} trend={trend} range={trendRange} onRange={setTrendRange}
            onAlert={async (id, target) => { await api.alert(id, target); notify('价格预警已建立，降价时会在站内通知'); }}
            onAddOffer={addOfferToOrder} addingOfferId={addingOfferId}
          />
          <PurchaseOrderPanel order={order} updatingItemId={updatingItemId} onQuantityChange={changeOrderQuantity} />
          <BudgetCalculator onSubmit={async (room, area) => {
            const result = await api.budget(room, area);
            notify('预算已保存，可继续替换为实际报价');
            return result.Estimate;
          }} />
        </>
      )}
      {message && <div className="toast" role="status">{message}</div>}
      <footer><span>筑价 BUILD PRICE INDEX</span><span>报价仅作采购决策参考，请以商家最终合同为准。</span></footer>
    </main>
  );
}
