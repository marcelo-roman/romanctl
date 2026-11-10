export interface ApiOrderItem {
  id: string;
  productName: string;
  quantity: number;
  unitPrice: number;
  subtotal: number;
}

export interface ApiOrder {
  id: string;
  customerName: string;
  status: string;
  createdAt: string;
  currency: string;
  total: number;
  canAddItems: boolean;
  canPlace: boolean;
  items: ApiOrderItem[];
}

export interface ApiCatalogProduct {
  id: string;
  name: string;
  price: number;
  currency: string;
}

export interface ApiNewOrderItem {
  productName: string;
  quantity: number;
  unitPrice: number;
  currency: string;
}
