class Item {
  final int id;
  final String name;
  final String desc;

  Item({required this.id, required this.name, required this.desc});

  factory Item.fromJson(Map<String, dynamic> json) {
    return Item(id: json['id'], name: json['name'], desc: json['desc']);
  }

  Map<String, dynamic> toJson() => {'id': id, 'name': name, 'desc': desc};
}
