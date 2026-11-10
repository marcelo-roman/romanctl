use std::fs;
use std::path::{Path, PathBuf};

const ROOT: &str = env!("CARGO_MANIFEST_DIR");
const LAYERS: [&str; 3] = ["domain", "application", "adapters"];

fn inner_layers(layer: &str) -> &'static [&'static str] {
    match layer {
        "domain" => &["domain"],
        "application" => &["domain", "application"],
        _ => &["domain", "application", "adapters"],
    }
}

fn permitted_crates(layer: &str) -> Option<&'static [&'static str]> {
    match layer {
        "domain" => Some(&["chrono", "thiserror", "uuid"]),
        "application" => Some(&["async_trait", "chrono", "thiserror", "uuid"]),
        _ => None,
    }
}

fn is_identifier(c: char) -> bool {
    c.is_ascii_alphanumeric() || c == '_'
}

fn without_comments(source: &str) -> String {
    source
        .lines()
        .map(|line| line.split("//").next().unwrap_or_default())
        .collect::<Vec<_>>()
        .join("\n")
}

fn leading_identifier(text: &str) -> &str {
    let text = text.trim_start();
    let end = text.find(|c| !is_identifier(c)).unwrap_or(text.len());
    &text[..end]
}

fn path_roots(source: &str) -> Vec<&str> {
    let mut roots = Vec::new();
    let mut previous = ' ';
    for (index, c) in source.char_indices() {
        let starts = is_identifier(c) && !is_identifier(previous) && previous != ':';
        previous = c;
        if !starts {
            continue;
        }
        let identifier = leading_identifier(&source[index..]);
        if source[index + identifier.len()..].starts_with("::") {
            roots.push(identifier);
        }
    }
    roots
}

fn crate_modules(source: &str) -> Vec<&str> {
    let mut modules = Vec::new();
    for (index, _) in source.match_indices("crate::") {
        let before = source[..index].chars().next_back().unwrap_or(' ');
        if is_identifier(before) || before == ':' {
            continue;
        }
        let rest = &source[index + "crate::".len()..];
        let Some(group) = rest.strip_prefix('{') else {
            modules.push(leading_identifier(rest));
            continue;
        };
        let mut depth = 0;
        let mut start = 0;
        for (offset, c) in group.char_indices() {
            match c {
                '{' => depth += 1,
                '}' if depth == 0 => {
                    modules.push(leading_identifier(&group[start..offset]));
                    break;
                }
                '}' => depth -= 1,
                ',' if depth == 0 => {
                    modules.push(leading_identifier(&group[start..offset]));
                    start = offset + 1;
                }
                _ => {}
            }
        }
    }
    modules.retain(|module| !module.is_empty());
    modules
}

fn dependencies() -> Vec<String> {
    let manifest = fs::read_to_string(Path::new(ROOT).join("Cargo.toml")).unwrap();
    manifest
        .lines()
        .skip_while(|line| line.trim() != "[dependencies]")
        .skip(1)
        .take_while(|line| !line.trim_start().starts_with('['))
        .filter_map(|line| line.split_once('='))
        .map(|(name, _)| name.trim().replace('-', "_"))
        .collect()
}

fn violations(layer: &str, file: &str, source: &str, dependencies: &[String]) -> Vec<String> {
    let source = without_comments(source);
    let mut found = Vec::new();
    for module in crate_modules(&source) {
        if LAYERS.contains(&module) && !inner_layers(layer).contains(&module) {
            found.push(format!("{file}: {layer} must not depend on {module}"));
        }
    }
    let Some(permitted) = permitted_crates(layer) else {
        return found;
    };
    let mut roots = path_roots(&source);
    roots.sort_unstable();
    roots.dedup();
    for root in roots {
        let external = dependencies.iter().any(|dependency| dependency == root);
        if external && !permitted.contains(&root) {
            found.push(format!("{file}: {layer} must not depend on {root}"));
        }
    }
    found
}

fn rust_files(directory: &Path) -> Vec<PathBuf> {
    let mut files = Vec::new();
    for entry in fs::read_dir(directory).unwrap() {
        let path = entry.unwrap().path();
        if path.is_dir() {
            files.extend(rust_files(&path));
        } else if path.extension().is_some_and(|extension| extension == "rs") {
            files.push(path);
        }
    }
    files
}

#[test]
fn dependencies_point_inward() {
    let dependencies = dependencies();
    assert!(dependencies.iter().any(|dependency| dependency == "axum"));
    let mut found = Vec::new();
    for layer in LAYERS {
        for path in rust_files(&Path::new(ROOT).join("src").join(layer)) {
            let source = fs::read_to_string(&path).unwrap();
            let file = path.strip_prefix(ROOT).unwrap().display().to_string();
            found.extend(violations(layer, &file, &source, &dependencies));
        }
    }
    assert!(found.is_empty(), "{found:#?}");
}

#[test]
fn outward_dependencies_are_reported() {
    let dependencies = vec!["axum".to_owned(), "chrono".to_owned()];
    let domain =
        "use axum::Json;\nuse chrono::Utc;\nuse crate::application::orders::OrderService;\n";
    assert_eq!(
        violations("domain", "model.rs", domain, &dependencies),
        vec![
            "model.rs: domain must not depend on application",
            "model.rs: domain must not depend on axum",
        ]
    );
    let application = "use crate::{domain::orders::Order, adapters::http::Services};\n";
    assert_eq!(
        violations("application", "service.rs", application, &dependencies),
        vec!["service.rs: application must not depend on adapters"]
    );
}
