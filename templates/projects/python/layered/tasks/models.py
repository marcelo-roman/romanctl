from django.db import models

TITLE_MAX_LENGTH = 200


class Task(models.Model):
    title = models.CharField(max_length=TITLE_MAX_LENGTH)
    done = models.BooleanField(default=False)
    due_date = models.DateField(null=True, blank=True)
    created_at = models.DateTimeField(auto_now_add=True)

    class Meta:
        ordering = ["-created_at", "-id"]

    def __str__(self) -> str:
        return self.title
