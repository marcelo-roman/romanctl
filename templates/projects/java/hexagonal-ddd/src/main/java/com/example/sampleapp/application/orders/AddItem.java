package com.example.sampleapp.application.orders;

import java.math.BigDecimal;

public record AddItem(String productName, int quantity, BigDecimal unitPrice, String currency) {
}
