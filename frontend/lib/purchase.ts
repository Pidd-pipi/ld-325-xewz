import type { Offer } from './types';

export function offerPurchasable(offer: Offer): boolean {
  return offer.StockStatus === 'in_stock' && offer.Supplier.Status === 'approved';
}

export function offerStockLabel(status: string): string {
  if (status === 'in_stock') return '有货';
  if (status === 'out_of_stock') return '缺货';
  return '停售';
}

export function supplierStatusLabel(status: string): string {
  if (status === 'approved') return '资质已通过';
  if (status === 'rejected') return '资质已驳回';
  return '资质未通过';
}

export function blockedReasonLabel(reason?: string): string {
  if (reason === 'offer_unavailable') return '库存停售或当前无货';
  if (reason === 'supplier_not_approved') return '供应商资质未通过或已驳回';
  if (reason === 'below_moq') return '数量低于起订量';
  return '不可采购';
}

export function signedMoney(value: number): string {
  if (value > 0) return `+${value.toFixed(2)}`;
  return value.toFixed(2);
}
