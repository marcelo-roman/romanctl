#[derive(Debug, Clone, Copy, Default)]
pub struct ListSampleItems {
    pub limit: Option<usize>,
}

impl ListSampleItems {
    pub const DEFAULT_LIMIT: usize = 50;
    pub const MAX_LIMIT: usize = 200;

    pub fn limit(&self) -> usize {
        self.limit
            .unwrap_or(Self::DEFAULT_LIMIT)
            .clamp(1, Self::MAX_LIMIT)
    }
}
