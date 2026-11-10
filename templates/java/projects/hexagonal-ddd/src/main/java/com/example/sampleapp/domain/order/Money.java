package com.example.sampleapp.domain.order;

import com.example.sampleapp.domain.shared.BusinessRuleException;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.Locale;

public record Money(BigDecimal amount, String currency) {

    public static final String DEFAULT_CURRENCY = "BRL";

    public Money {
        if (amount == null) {
            throw new BusinessRuleException("O valor é obrigatório.");
        }
        if (amount.signum() < 0) {
            throw new BusinessRuleException("O valor não pode ser negativo.");
        }
        if (currency == null || currency.strip().length() != 3) {
            throw new BusinessRuleException("A moeda deve ter 3 letras, como BRL.");
        }
        amount = amount.setScale(2, RoundingMode.HALF_EVEN);
        currency = currency.strip().toUpperCase(Locale.ROOT);
    }

    public static Money of(BigDecimal amount, String currency) {
        if (currency == null || currency.isEmpty()) {
            return new Money(amount, DEFAULT_CURRENCY);
        }
        return new Money(amount, currency);
    }

    public static Money zero(String currency) {
        return new Money(BigDecimal.ZERO, currency);
    }

    public Money multiply(int quantity) {
        return new Money(amount.multiply(BigDecimal.valueOf(quantity)), currency);
    }

    public Money add(Money other) {
        if (!other.currency.equals(currency)) {
            throw new BusinessRuleException("Não é possível somar moedas diferentes.");
        }
        return new Money(amount.add(other.amount), currency);
    }
}
