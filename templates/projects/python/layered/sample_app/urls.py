from django.urls import include, path

from sample_app.views import health

urlpatterns = [
    path("health", health, name="health"),
    path("", include("tasks.urls")),
    # romanctl:urls
]

handler404 = "sample_app.problems.page_not_found"
handler500 = "sample_app.problems.server_error"
