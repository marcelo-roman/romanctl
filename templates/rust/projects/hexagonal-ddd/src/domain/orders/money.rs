use crate::domain::shared::RuleViolation;

pub const DEFAULT_CURRENCY: &str = "USD";

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Money {
    cents: i64,
    currency: String,
}

impl Money {
    pub fn new(cents: i64, currency: &str) -> Result<Self, RuleViolation> {
        if cents < 0 {
            return Err(RuleViolation::new("Amount cannot be negative."));
        }
        let currency = currency.trim();
        if currency.len() != 3 || !currency.chars().all(|c| c.is_ascii_alphabetic()) {
            return Err(RuleViolation::new(
                "Currency must have 3 letters, like USD.",
            ));
        }
        Ok(Self {
            cents,
            currency: currency.to_ascii_uppercase(),
        })
    }

    pub(super) fn trusted(cents: i64, currency: &str) -> Self {
        Self {
            cents,
            currency: currency.to_owned(),
        }
    }

    pub fn from_decimal(amount: f64, currency: &str) -> Result<Self, RuleViolation> {
        if !amount.is_finite() {
            return Err(RuleViolation::new("The amount provided is invalid."));
        }
        Self::new((amount * 100.0).round() as i64, currency)
    }

    pub fn cents(&self) -> i64 {
        self.cents
    }

    pub fn amount(&self) -> f64 {
        self.cents as f64 / 100.0
    }

    pub fn currency(&self) -> &str {
        &self.currency
    }

    pub fn multiply(&self, quantity: i32) -> Self {
        Self {
            cents: self.cents * i64::from(quantity),
            currency: self.currency.clone(),
        }
    }

    pub fn add(&self, other: &Money) -> Result<Self, RuleViolation> {
        if other.currency != self.currency {
            return Err(RuleViolation::new("Cannot add different currencies."));
        }
        Ok(Self {
            cents: self.cents + other.cents,
            currency: self.currency.clone(),
        })
    }
}
