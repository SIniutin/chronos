import 'package:flutter/foundation.dart';

class AppNavigation {
  static final mainTab = ValueNotifier<int>(0);

  static void goToTab(int index) {
    mainTab.value = index;
  }

  static void goHome() {
    goToTab(0);
  }
}
