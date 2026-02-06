import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../app_state.dart';
import '../services/api_client.dart';

class AuditLogsScreen extends StatefulWidget {
  const AuditLogsScreen({super.key});

  @override
  State<AuditLogsScreen> createState() => _AuditLogsScreenState();
}

class _AuditLogsScreenState extends State<AuditLogsScreen> {
  Future<List<AuditLogItem>>? _future;
  ApiClient? _client;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final client = context.read<AppState>().apiClient;
    if (client?.baseUrl != _client?.baseUrl ||
        client?.apiKey != _client?.apiKey ||
        client?.token != _client?.token) {
      _client = client;
      if (client != null) {
        _future = _load(client);
      }
    }
  }

  Future<List<AuditLogItem>> _load(ApiClient client) async {
    final resp = await client.getJson('/admin/api/v1/audit-logs');
    final items = (resp['items'] as List<dynamic>? ?? [])
        .map((e) => AuditLogItem.fromJson(e as Map<String, dynamic>))
        .toList();
    return items;
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<AuditLogItem>>(
      future: _future,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Scaffold(
            body: Center(child: CircularProgressIndicator()),
          );
        }
        if (snapshot.hasError) {
          return Scaffold(
            appBar: AppBar(
              title: const Text('操作日志'),
              actions: [
                IconButton(
                  icon: const Icon(Icons.refresh),
                  onPressed: () {
                    final client = context.read<AppState>().apiClient;
                    if (client != null) {
                      setState(() {
                        _future = _load(client);
                      });
                    }
                  },
                ),
              ],
            ),
            body: _ErrorState(
              message: '加载日志失败，请检查 API Key 权限。',
              onRetry: () {
                final client = context.read<AppState>().apiClient;
                if (client != null) {
                  setState(() {
                    _future = _load(client);
                  });
                }
              },
            ),
          );
        }
        final items = snapshot.data ?? [];
        return Scaffold(
          appBar: AppBar(
            title: const Text('操作日志'),
            actions: [
              IconButton(
                icon: const Icon(Icons.refresh),
                onPressed: () {
                  final client = context.read<AppState>().apiClient;
                  if (client != null) {
                    setState(() {
                      _future = _load(client);
                    });
                  }
                },
              ),
            ],
          ),
          body: RefreshIndicator(
            onRefresh: () async {
              final client = context.read<AppState>().apiClient;
              if (client != null) {
                setState(() {
                  _future = _load(client);
                });
              }
              await _future;
            },
            child: ListView.builder(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 24),
              itemCount: items.isEmpty ? 1 : items.length,
              itemBuilder: (context, index) {
                if (items.isEmpty) {
                  return const _EmptyState();
                }
                final item = items[index];
                final theme = Theme.of(context);
                final colorScheme = theme.colorScheme;
                return Container(
                  margin: const EdgeInsets.only(bottom: 12),
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: colorScheme.surface,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: colorScheme.outlineVariant.withOpacity(0.5),
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: colorScheme.shadow.withOpacity(0.05),
                        blurRadius: 8,
                        offset: const Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: colorScheme.primary.withOpacity(0.12),
                          borderRadius: BorderRadius.circular(10),
                        ),
                        child: Icon(Icons.history, color: colorScheme.primary),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              item.action,
                              style: theme.textTheme.titleSmall?.copyWith(
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Text(
                              '${item.targetType} · ${item.targetId}',
                              style: theme.textTheme.bodySmall?.copyWith(
                                color: colorScheme.onSurfaceVariant,
                              ),
                            ),
                          ],
                        ),
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: colorScheme.primaryContainer.withOpacity(0.5),
                          borderRadius: BorderRadius.circular(999),
                        ),
                        child: Text(
                          item.createdAt,
                          style: theme.textTheme.labelSmall?.copyWith(
                            color: colorScheme.primary,
                          ),
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),
        );
      },
    );
  }
}

class AuditLogItem {
  final int id;
  final String action;
  final String targetType;
  final String targetId;
  final String createdAt;

  AuditLogItem({
    required this.id,
    required this.action,
    required this.targetType,
    required this.targetId,
    required this.createdAt,
  });

  factory AuditLogItem.fromJson(Map<String, dynamic> json) {
    return AuditLogItem(
      id: json['id'] as int? ?? 0,
      action: json['action'] as String? ?? '未知操作',
      targetType: json['target_type'] as String? ?? '',
      targetId: json['target_id'] as String? ?? '',
      createdAt: json['created_at'] as String? ?? '',
    );
  }
}

class _ErrorState extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;

  const _ErrorState({required this.message, required this.onRetry});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 16),
            FilledButton(onPressed: onRetry, child: const Text('重试')),
          ],
        ),
      ),
    );
  }
}

class _EmptyState extends StatelessWidget {
  const _EmptyState();

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.all(24),
      child: Center(child: Text('暂无日志')),
    );
  }
}
