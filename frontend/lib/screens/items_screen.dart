import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

import '../config/constants.dart';
import '../models/item_model.dart';

class ItemsScreen extends StatefulWidget {
  const ItemsScreen({super.key});

  @override
  State<ItemsScreen> createState() => _ItemsScreenState();
}

class _ItemsScreenState extends State<ItemsScreen> {
  List<Item> _items = [];
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    getItems();
  }

  void getItems() async {
    setState(() {
      _isLoading = true;
    });
    final response = await http.get(Uri.parse(apiUrl));
    if (response.statusCode == 200) {
      final List<dynamic> data = jsonDecode(response.body);
      final List<Item> items = data.map((datadariapi) => Item.fromJson(datadariapi)).toList();

      setState(() {
        _items = items;
        _isLoading = false;
      });
    }
  }

  void createItem(String name, String desc) async {
    final response = await http.post(Uri.parse(apiUrl), body: jsonEncode({'name': name, 'desc': desc}));

    if (response.statusCode == 201) {
      final newItem = Item.fromJson(jsonDecode(response.body));
      setState(() {
        _items.add(newItem);
      });
    }
  }

  void _showItemDialog([Item? item]) {
    final nameCtrl = TextEditingController(text: item?.name ?? '');
    final descCtrl = TextEditingController(text: item?.desc ?? '');

    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: Text(item == null ? 'New Item' : 'Edit Item'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: nameCtrl,
                decoration: const InputDecoration(labelText: 'Name'),
              ),
              TextField(
                controller: descCtrl,
                decoration: const InputDecoration(labelText: 'Description'),
              ),
            ],
          ),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
            ElevatedButton(
              onPressed: () {
                createItem(nameCtrl.text, descCtrl.text);
                Navigator.pop(context);
              },
              child: const Text('Save'),
            ),
          ],
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('CRUD Items'),
        actions: [IconButton(icon: const Icon(Icons.refresh), onPressed: null)],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _items.isEmpty
          ? const Center(child: Text("No items found. Create one!"))
          : ListView.builder(
              itemCount: _items.length,
              itemBuilder: (context, index) {
                final item = _items[index];
                return ListTile(
                  title: Text(item.name),
                  subtitle: Text(item.desc),
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        icon: const Icon(Icons.edit, color: Colors.blue),
                        onPressed: () => _showItemDialog(item),
                      ),
                      IconButton(
                        icon: const Icon(Icons.delete, color: Colors.red),
                        onPressed: () => null,
                      ),
                    ],
                  ),
                );
              },
            ),
      floatingActionButton: FloatingActionButton(onPressed: () => _showItemDialog(), child: const Icon(Icons.add)),
    );
  }
}
