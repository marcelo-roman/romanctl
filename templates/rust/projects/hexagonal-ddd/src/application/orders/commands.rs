#[derive(Debug, Clone)]
pub struct AddItem {
    pub product_name: String,
    pub quantity: i32,
    pub unit_price: f64,
    pub currency: Option<String>,
}

#[derive(Debug, Clone, Copy, Default)]
pub struct ListOrders {
    pub limit: Option<usize>,
}

impl ListOrders {
    pub const DEFAULT_LIMIT: usize = 50;
    pub const MAX_LIMIT: usize = 200;

    pub fn limit(&self) -> usize {
        self.limit
            .unwrap_or(Self::DEFAULT_LIMIT)
            .clamp(1, Self::MAX_LIMIT)
    }
}
