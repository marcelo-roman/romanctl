package com.example.sampleapp.application.shared;

public final class ListLimit {

    public static final int DEFAULT = 50;
    public static final int MAX = 200;

    private ListLimit() {
    }

    public static int of(Integer requested) {
        if (requested == null || requested <= 0) {
            return DEFAULT;
        }
        return Math.min(requested, MAX);
    }
}
