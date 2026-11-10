package com.example.sampleapp.architecture;

import static com.tngtech.archunit.lang.syntax.ArchRuleDefinition.classes;
import static com.tngtech.archunit.lang.syntax.ArchRuleDefinition.noClasses;
import static com.tngtech.archunit.library.Architectures.layeredArchitecture;

import com.tngtech.archunit.core.importer.ImportOption;
import com.tngtech.archunit.junit.AnalyzeClasses;
import com.tngtech.archunit.junit.ArchTest;
import com.tngtech.archunit.lang.ArchRule;

@AnalyzeClasses(packages = "com.example.sampleapp", importOptions = ImportOption.DoNotIncludeTests.class)
class ArchitectureTest {

    private static final String ROOT = "com.example.sampleapp";

    @ArchTest
    static final ArchRule domainDependsOnNothing = classes().that().resideInAPackage(ROOT + ".domain..")
            .should().onlyDependOnClassesThat().resideInAnyPackage(ROOT + ".domain..", "java..");

    @ArchTest
    static final ArchRule applicationDependsOnlyOnDomain = classes().that().resideInAPackage(ROOT + ".application..")
            .should().onlyDependOnClassesThat().resideInAnyPackage(ROOT + ".domain..", ROOT + ".application..", "java..");

    @ArchTest
    static final ArchRule adaptersDoNotDependOnConfiguration = noClasses().that().resideInAPackage(ROOT + ".adapters..")
            .should().dependOnClassesThat().resideInAPackage(ROOT + ".config..");

    @ArchTest
    static final ArchRule dependenciesPointInward = layeredArchitecture().consideringOnlyDependenciesInLayers()
            .layer("Domain").definedBy(ROOT + ".domain..")
            .layer("Application").definedBy(ROOT + ".application..")
            .layer("Adapters").definedBy(ROOT + ".adapters..")
            .layer("Config").definedBy(ROOT + ".config..")
            .whereLayer("Config").mayNotBeAccessedByAnyLayer()
            .whereLayer("Adapters").mayOnlyBeAccessedByLayers("Config")
            .whereLayer("Application").mayOnlyBeAccessedByLayers("Adapters", "Config")
            .whereLayer("Domain").mayOnlyBeAccessedByLayers("Application", "Adapters", "Config");
}
