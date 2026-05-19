import 'dart:io';
import 'package:flutter/foundation.dart';

String get backendHost {
  if (kIsWeb) return 'localhost';
  if (Platform.isAndroid) return '10.0.2.2';
  return 'localhost';
}

final String apiUrl = 'http://$backendHost:8080/api/items';
final String wsUrl = 'ws://$backendHost:8080/ws/chat';
