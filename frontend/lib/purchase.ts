import type { Offer, PurchaseOrderItem } from './types';

export function isOfferPurchasable(offer: Offer): boolean {
  return offer.StockStatus === 'in_stock' && offer.Supplier.Status === 'approved';
}

export function offerBlockReason(offer: Offer): string {
  if (offer.StockStatus !== 'in_stock') return offer.StockStatus === 'discontinued' ? '已停售' : '缺货';
  if (offer.Supplier.Status === 'rejected') return '资质已驳回';
  return '资质待审核';
}

export function itemBlockReason(item: PurchaseOrderItem): string {
  if (item.stock_status !== 'in_stock') return item.stock_status === 'discontinued' ? '库存停售' : '库存不足';
  if (item.supplier_status === 'rejected') return '资质驳回';
  return '资质待审核';
}
