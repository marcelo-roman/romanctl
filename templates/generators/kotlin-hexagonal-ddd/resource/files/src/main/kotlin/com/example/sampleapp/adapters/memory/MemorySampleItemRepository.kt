package com.example.sampleapp.adapters.memory

import com.example.sampleapp.application.sampleitems.SampleItemRepository
import com.example.sampleapp.domain.sampleitem.SampleItem
import java.util.UUID
import java.util.concurrent.ConcurrentHashMap

class MemorySampleItemRepository : SampleItemRepository {
    private val sampleItems = ConcurrentHashMap<UUID, SampleItem>()

    override fun save(sampleItem: SampleItem) {
        sampleItems[sampleItem.id] = copy(sampleItem)
    }

    override fun findById(id: UUID): SampleItem? = sampleItems[id]?.let(::copy)

    override fun list(limit: Int): List<SampleItem> =
        sampleItems.values.sortedByDescending { it.createdAt }.take(limit).map(::copy)

    override fun delete(id: UUID): Boolean = sampleItems.remove(id) != null

    private fun copy(sampleItem: SampleItem): SampleItem =
        SampleItem.restore(sampleItem.id, sampleItem.name, sampleItem.createdAt)
}
