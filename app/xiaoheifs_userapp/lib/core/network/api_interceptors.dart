import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import '../constants/api_endpoints.dart';
import '../navigation/app_navigator.dart';
import '../storage/storage_service.dart';
import 'api_client.dart';

class AuthInterceptor extends Interceptor {
  final StorageService _storage = StorageService.instance;
  static Future<String?>? _refreshing;
  static bool _unauthorizedHandled = false;

  @override
  void onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = _storage.getAccessToken();
    if (token != null && token.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $token';
      _unauthorizedHandled = false;
    }

    if (options.headers.containsKey('X-Use-Api-Key')) {
      options.headers.remove('X-Use-Api-Key');
    }

    handler.next(options);
  }

  @override
  void onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    if (err.response?.statusCode == 401 && err.requestOptions.extra['retried'] != true) {
      if (_unauthorizedHandled) {
        handler.next(err);
        return;
      }
      final path = err.requestOptions.path;
      final isAuthEndpoint =
          path.contains(ApiEndpoints.authLogin) || path.contains(ApiEndpoints.authRefresh);

      if (!isAuthEndpoint) {
        final refreshToken = _storage.getRefreshToken();
        if (refreshToken != null && refreshToken.isNotEmpty) {
          try {
            _refreshing ??= _refreshToken(refreshToken);
            final newToken = await _refreshing;
            _refreshing = null;

            if (newToken != null && newToken.isNotEmpty) {
              _unauthorizedHandled = false;
              final options = err.requestOptions;
              options.headers['Authorization'] = 'Bearer $newToken';
              options.extra['retried'] = true;
              try {
                final response = await ApiClient.instance.dio.fetch(options);
                handler.resolve(response);
                return;
              } catch (_) {}
            }
          } finally {
            _refreshing = null;
          }
        }
      }

      _unauthorizedHandled = true;
      await _storage.clearAuthData();
      AppNavigator.showSnackBar('鉴权失败，请重新登录');
      AppNavigator.goToLogin();
    }

    handler.next(err);
  }

  Future<String?> _refreshToken(String refreshToken) async {
    try {
      final baseUrl = ApiClient.instance.dio.options.baseUrl;
      final dio = Dio(BaseOptions(baseUrl: baseUrl));
      final response = await dio.post(
        ApiEndpoints.authRefresh,
        data: {'refresh_token': refreshToken},
      );
      final data = response.data is Map ? response.data as Map : {};
      final access = data['access_token'] ?? data['accessToken'] ?? data['token'];
      final newRefresh = data['refresh_token'] ?? data['refreshToken'];
      if (access is String && access.isNotEmpty) {
        await _storage.setAccessToken(access);
        if (newRefresh is String && newRefresh.isNotEmpty) {
          await _storage.setRefreshToken(newRefresh);
        }
        return access;
      }
    } catch (_) {}
    return null;
  }
}

class ErrorInterceptor extends Interceptor {
  static bool _realnameDialogOpen = false;

  @override
  void onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) {
    final status = err.response?.statusCode;
    final data = err.response?.data;
    final message = _extractMessage(data) ?? err.message ?? 'Request failed';
    final url = err.requestOptions.path;

    if (status == 403 && url.startsWith('/v1') && message.toLowerCase().contains('real name required')) {
      if (!_realnameDialogOpen) {
        _realnameDialogOpen = true;
        AppNavigator.showConfirmDialog(
          title: '需要实名认证',
          content: '该操作需要完成实名认证，是否前往认证页面？',
          confirmText: '去认证',
          cancelText: '稍后再说',
        ).then((confirmed) {
          _realnameDialogOpen = false;
          if (confirmed == true) {
            AppNavigator.goToRealname();
          }
        });
      }
      handler.next(err);
      return;
    }

    if (status != null && status >= 500) {
      AppNavigator.showSnackBar('服务端错误：$message', backgroundColor: Colors.red);
    } else if (status != null) {
      AppNavigator.showSnackBar(message, backgroundColor: Colors.red);
    }

    handler.next(DioException(
      requestOptions: err.requestOptions,
      response: err.response,
      type: err.type,
      error: message,
      message: message,
    ));
  }

  String? _extractMessage(dynamic data) {
    if (data is Map<String, dynamic>) {
      return data['error']?.toString() ?? data['message']?.toString();
    }
    return null;
  }
}

class RetryInterceptor extends Interceptor {
  final int maxRetries;
  final Dio dio;
  int retryCount = 0;

  RetryInterceptor({
    required this.dio,
    this.maxRetries = 3,
  });

  @override
  void onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    if (_shouldRetry(err)) {
      if (retryCount < maxRetries) {
        retryCount++;
        try {
          final response = await dio.fetch(err.requestOptions);
          retryCount = 0;
          handler.resolve(response);
          return;
        } catch (_) {}
      }
      retryCount = 0;
    }

    handler.next(err);
  }

  bool _shouldRetry(DioException error) {
    return error.type == DioExceptionType.connectionError ||
        (error.type == DioExceptionType.badResponse &&
            error.response?.statusCode != null &&
            error.response!.statusCode! >= 500);
  }
}
