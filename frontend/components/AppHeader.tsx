'use client';

import { Bell, Calculator, Scale, Search, ShoppingCart } from 'lucide-react';

export function AppHeader({ query, onQuery, orderCount }: { query: string; onQuery(value: string): void; orderCount: number }) {
  return (
    <header className="site-header">
      <a className="brand" href="#top">
        <span className="brand-mark">筑</span>
        <span>筑价<small>BUILD PRICE INDEX</small></span>
      </a>
      <nav>
        <a href="#catalog">找材料</a>
        <a href="#compare">比报价</a>
        <a href="#purchase-order">采购单</a>
        <a href="#budget">算预算</a>
      </nav>
      <label className="search">
        <Search size={17} />
        <input value={query} onChange={(event) => onQuery(event.target.value)} placeholder="搜索品牌、型号或建材" />
      </label>
      <div className="header-actions">
        <span title="对比清单"><Scale size={18} /></span>
        <a className="cart-link" href="#purchase-order" title="锁价采购单">
          <ShoppingCart size={18} />
          {orderCount > 0 && <b>{orderCount}</b>}
        </a>
        <span title="价格预警"><Bell size={18} /></span>
        <span title="预算工具"><Calculator size={18} /></span>
      </div>
    </header>
  );
}
