use chrono::Utc;
use sample_app::domain::sample_items::SampleItem;

#[test]
fn create_requires_name() {
    assert!(SampleItem::create(" ", Utc::now()).is_err());
    let sample_item = SampleItem::create("  First ", Utc::now()).unwrap();
    assert_eq!(sample_item.name(), "First");
}

#[test]
fn rename_keeps_the_rule() {
    let mut sample_item = SampleItem::create("First", Utc::now()).unwrap();
    assert!(sample_item.rename("").is_err());
    assert_eq!(sample_item.name(), "First");
    sample_item.rename("Second").unwrap();
    assert_eq!(sample_item.name(), "Second");
}
