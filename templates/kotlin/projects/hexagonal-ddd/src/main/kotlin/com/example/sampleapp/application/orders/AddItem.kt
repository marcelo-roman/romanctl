package com.example.sampleapp.application.orders

import java.math.BigDecimal

data class AddItem(val productName: String, val quantity: Int, val unitPrice: BigDecimal?, val currency: String?)
