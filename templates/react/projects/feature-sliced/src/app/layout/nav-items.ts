import { paths } from '@/shared/config';

export interface NavItem {
  to: string;
  label: string;
}

export const navItems: NavItem[] = [
  { to: paths.orders, label: 'Pedidos' },
  // romanctl:nav
];
