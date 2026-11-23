package com.example.sampleapp.domain.order

import com.example.sampleapp.domain.shared.BusinessRuleException
import java.math.BigDecimal
import java.math.RoundingMode

class Money(amount: BigDecimal, currency: String) {
    val amount: BigDecimal
    val currency: String

    init {
        if (amount.signum() < 0) {
            throw BusinessRuleException("Amount cannot be negative.")
        }
        if (!CURRENCY.matches(currency.trim())) {
            throw BusinessRuleException("Currency must have 3 letters, like BRL.")
        }
        this.amount = amount.setScale(2, RoundingMode.HALF_EVEN)
        this.currency = currency.trim().uppercase()
    }

    operator fun times(quantity: Int): Money = Money(amount.multiply(BigDecimal.valueOf(quantity.toLong())), currency)

    operator fun plus(other: Money): Money {
        if (other.currency != currency) {
            throw BusinessRuleException("Cannot add different currencies.")
        }
        return Money(amount.add(other.amount), currency)
    }

    override fun equals(other: Any?): Boolean = other is Money && other.amount == amount && other.currency == currency

    override fun hashCode(): Int = 31 * amount.hashCode() + currency.hashCode()

    override fun toString(): String = "${amount.toPlainString()} $currency"

    companion object {
        const val DEFAULT_CURRENCY = "BRL"
        private val CURRENCY = Regex("^[A-Za-z]{3}$")

        fun of(amount: BigDecimal?, currency: String?): Money {
            if (amount == null) {
                throw BusinessRuleException("Amount is required.")
            }
            if (currency.isNullOrEmpty()) {
                return Money(amount, DEFAULT_CURRENCY)
            }
            return Money(amount, currency)
        }

        fun zero(currency: String): Money = Money(BigDecimal.ZERO, currency)
    }
}
