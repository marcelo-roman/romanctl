use super::Money;
use crate::domain::shared::{Id, RuleViolation};

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct OrderItem {
    id: Id,
    product_name: String,
    quantity: i32,
    unit_price: Money,
}

impl OrderItem {
    pub(super) fn create(
        product_name: &str,
        quantity: i32,
        unit_price: Money,
    ) -> Result<Self, RuleViolation> {
        let product_name = product_name.trim();
        if product_name.is_empty() {
            return Err(RuleViolation::new("O nome do produto é obrigatório."));
        }
        if quantity <= 0 {
            return Err(RuleViolation::new("A quantidade deve ser maior que zero."));
        }
        Ok(Self {
            id: Id::generate(),
            product_name: product_name.to_owned(),
            quantity,
            unit_price,
        })
    }

    pub fn restore(id: Id, product_name: String, quantity: i32, unit_price: Money) -> Self {
        Self {
            id,
            product_name,
            quantity,
            unit_price,
        }
    }

    pub fn id(&self) -> Id {
        self.id
    }

    pub fn product_name(&self) -> &str {
        &self.product_name
    }

    pub fn quantity(&self) -> i32 {
        self.quantity
    }

    pub fn unit_price(&self) -> &Money {
        &self.unit_price
    }

    pub fn subtotal(&self) -> Money {
        self.unit_price.multiply(self.quantity)
    }
}
