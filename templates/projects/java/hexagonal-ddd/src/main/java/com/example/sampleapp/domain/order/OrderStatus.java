package com.example.sampleapp.domain.order;

public enum OrderStatus {
    DRAFT("Draft"),
    PLACED("Placed");

    private final String label;

    OrderStatus(String label) {
        this.label = label;
    }

    public String label() {
        return label;
    }
}
