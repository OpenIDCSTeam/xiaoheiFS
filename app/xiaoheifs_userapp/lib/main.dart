import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/navigation/app_navigator.dart';
import 'core/network/api_client.dart';
import 'core/storage/storage_service.dart';
import 'presentation/routes/app_routes.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await StorageService.init();

  final client = ApiClient.instance;
  final savedBaseUrl = StorageService.instance.getApiBaseUrl();
  if (savedBaseUrl != null && savedBaseUrl.isNotEmpty) {
    client.updateBaseUrl(savedBaseUrl);
  }

  runApp(
    const ProviderScope(
      child: MyApp(),
    ),
  );
}

class MyApp extends ConsumerWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: '云享互联',
      debugShowCheckedModeBanner: false,
      scaffoldMessengerKey: AppNavigator.messengerKey,
      builder: (context, child) {
        final media = MediaQuery.of(context);
        final appChild = MediaQuery(
          data: media.copyWith(textScaler: const TextScaler.linear(1.0)),
          child: child ?? const SizedBox.shrink(),
        );

        if (defaultTargetPlatform != TargetPlatform.android) {
          return appChild;
        }

        return Align(
          alignment: Alignment.topCenter,
          child: Transform.scale(
            scale: 0.75,
            alignment: Alignment.topCenter,
            child: SizedBox(
              width: media.size.width / 0.75,
              height: media.size.height / 0.75,
              child: appChild,
            ),
          ),
        );
      },
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.blue,
          brightness: Brightness.light,
        ),
        visualDensity: VisualDensity.compact,
        materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
        useMaterial3: true,
      ),
      darkTheme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.blue,
          brightness: Brightness.dark,
        ),
        visualDensity: VisualDensity.compact,
        materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
        useMaterial3: true,
      ),
      themeMode: ThemeMode.system,
      routerConfig: router,
    );
  }
}
