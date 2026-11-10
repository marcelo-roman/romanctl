from django.db import models

NAME_MAX_LENGTH = 200


class SampleItem(models.Model):
    name = models.CharField(max_length=NAME_MAX_LENGTH)
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ["-created_at", "-id"]

    def __str__(self) -> str:
        return self.name
