package com.example.sampleapp.application.sampleitems

import com.example.sampleapp.domain.sampleitem.SampleItem
import java.util.UUID

interface SampleItemRepository {
    fun save(sampleItem: SampleItem)

    fun findById(id: UUID): SampleItem?

    fun list(limit: Int): List<SampleItem>

    fun delete(id: UUID): Boolean
}
