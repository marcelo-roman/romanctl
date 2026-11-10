from django.urls import path

from sample_items.views import SampleItemDetailView, SampleItemListView

urlpatterns = [
    path("sample-items", SampleItemListView.as_view(), name="sample-item-list"),
    path("sample-items/<int:sample_item_id>", SampleItemDetailView.as_view(), name="sample-item-detail"),
]
