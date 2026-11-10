package com.example.sampleapp.architecture

import com.lemonappdev.konsist.api.Konsist
import com.lemonappdev.konsist.api.architecture.KoArchitectureCreator.assertArchitecture
import com.lemonappdev.konsist.api.architecture.Layer
import com.lemonappdev.konsist.api.verify.assertTrue
import kotlin.test.Test

class ArchitectureTest {
    private val root = "com.example.sampleapp"

    @Test
    fun `dependencies point inward`() {
        Konsist.scopeFromProduction().assertArchitecture {
            val domain = Layer("Domain", "$root.domain..")
            val application = Layer("Application", "$root.application..")
            val adapters = Layer("Adapters", "$root.adapters..")

            domain.dependsOnNothing()
            application.dependsOn(domain)
            adapters.dependsOn(application, domain)
        }
    }

    @Test
    fun `domain and application know no framework`() {
        val allowed = listOf("java.", "kotlin.", "$root.domain.", "$root.application.")
        listOf("$root.domain..", "$root.application..").forEach { layer ->
            Konsist.scopeFromPackage(layer, sourceSetName = "main").imports.assertTrue { import ->
                allowed.any { import.name.startsWith(it) }
            }
        }
    }
}
