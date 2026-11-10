package com.example.sampleapp.application.sampleitems

import com.example.sampleapp.application.shared.ListLimit
import com.example.sampleapp.application.shared.NotFoundException
import com.example.sampleapp.domain.sampleitem.SampleItem
import com.example.sampleapp.domain.shared.Identifier
import java.time.Clock
import java.util.UUID

class SampleItemService(private val repository: SampleItemRepository, private val clock: Clock) {
    fun create(name: String): UUID {
        val sampleItem = SampleItem.create(name, clock.instant())
        repository.save(sampleItem)
        return sampleItem.id
    }

    fun rename(id: String, name: String) {
        val sampleItem = load(id)
        sampleItem.rename(name)
        repository.save(sampleItem)
    }

    fun get(id: String): SampleItemView = SampleItemView.of(load(id))

    fun list(limit: Int?): List<SampleItemView> = repository.list(ListLimit.of(limit)).map(SampleItemView::of)

    fun delete(id: String) {
        if (!repository.delete(Identifier.parse(id))) {
            throw notFound()
        }
    }

    private fun load(id: String): SampleItem = repository.findById(Identifier.parse(id)) ?: throw notFound()

    private fun notFound() = NotFoundException("SampleItem não encontrado.")
}
