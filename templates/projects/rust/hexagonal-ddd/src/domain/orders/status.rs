use crate::domain::shared::RuleViolation;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum OrderStatus {
    Draft,
    Placed,
}

impl OrderStatus {
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::Draft => "Draft",
            Self::Placed => "Placed",
        }
    }

    pub fn parse(value: &str) -> Result<Self, RuleViolation> {
        match value {
            "Draft" => Ok(Self::Draft),
            "Placed" => Ok(Self::Placed),
            _ => Err(RuleViolation::new(format!(
                "Situação de pedido desconhecida: {value}."
            ))),
        }
    }
}
