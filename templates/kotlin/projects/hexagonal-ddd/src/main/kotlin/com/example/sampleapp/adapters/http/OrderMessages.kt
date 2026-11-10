package com.example.sampleapp.adapters.http

import com.example.sampleapp.application.orders.AddItem
import com.example.sampleapp.application.orders.OrderItemView
import com.example.sampleapp.application.orders.OrderView
import java.math.BigDecimal
import kotlinx.serialization.Serializable

@Serializable
data class CreateOrderRequest(val customerName: String = "")

@Serializable
data class AddItemRequest(
    val productName: String = "",
    val quantity: Int = 0,
    @Serializable(with = BigDecimalSerializer::class) val unitPrice: BigDecimal? = null,
    val currency: String? = null,
) {
    fun toCommand(): AddItem = AddItem(productName, quantity, unitPrice, currency)
}

@Serializable
data class OrderItemResponse(
    val productName: String,
    val quantity: Int,
    @Serializable(with = BigDecimalSerializer::class) val unitPrice: BigDecimal,
    @Serializable(with = BigDecimalSerializer::class) val subtotal: BigDecimal,
) {
    companion object {
        fun of(view: OrderItemView): OrderItemResponse =
            OrderItemResponse(view.productName, view.quantity, view.unitPrice, view.subtotal)
    }
}

@Serializable
data class OrderResponse(
    val id: String,
    val customerName: String,
    val status: String,
    val createdAt: String,
    @Serializable(with = BigDecimalSerializer::class) val total: BigDecimal,
    val currency: String,
    val items: List<OrderItemResponse>,
) {
    companion object {
        fun of(view: OrderView): OrderResponse = OrderResponse(
            view.id.toString(),
            view.customerName,
            view.status,
            view.createdAt.toString(),
            view.total,
            view.currency,
            view.items.map(OrderItemResponse::of),
        )
    }
}
