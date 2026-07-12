import 'package:flutter/material.dart';
import '../../core/theme/colors.dart';

class PrivacyPage extends StatelessWidget {
  const PrivacyPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('隐私政策'),
        backgroundColor: AppColors.darkBg,
        foregroundColor: AppColors.textPrimary,
      ),
      backgroundColor: AppColors.darkBg,
      body: const SingleChildScrollView(
        padding: EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '隐私政策',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            SizedBox(height: 8),
            Text(
              '更新日期：2026年6月9日',
              style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
            ),
            SizedBox(height: 16),
            _Section(
              title: '一、信息收集',
              content: '1.1 我们收集的信息包括：\n'
                  '（1）注册信息：手机号码、密码（加密存储）；\n'
                  '（2）使用数据：您在使用本服务过程中产生的对话记录、情绪记录、反思笔记等内容；\n'
                  '（3）设备信息：设备型号、操作系统版本等基础信息。\n\n'
                  '1.2 我们仅在提供本服务所必需的范围内收集上述信息。',
            ),
            _Section(
              title: '二、信息使用',
              content: '2.1 我们使用收集的信息用于以下目的：\n'
                  '（1）创建和管理您的账号；\n'
                  '（2）提供个性化的 AI 对话与自我探索体验；\n'
                  '（3）优化和改进本服务的功能与性能；\n'
                  '（4）保障账号安全和防范欺诈。\n\n'
                  '2.2 我们不会将您的个人信息用于上述目的之外的用途，除非获得您的明确同意或法律要求。',
            ),
            _Section(
              title: '三、信息存储与保护',
              content: '3.1 您的数据存储在位于中华人民共和国的服务器上。\n\n'
                  '3.2 我们采用业界通行的安全技术（包括数据加密传输、访问控制、安全审计等）保护您的个人信息。\n\n'
                  '3.3 您的密码采用 bcrypt 哈希算法加密存储，我们无法获知您的明文密码。',
            ),
            _Section(
              title: '四、信息共享与披露',
              content: '4.1 我们承诺不会向任何第三方出售、出租或交易您的个人信息。\n\n'
                  '4.2 在以下情况下，我们可能共享必要的信息：\n'
                  '（1）获得您的明确同意；\n'
                  '（2）为完成您所请求的服务（如短信验证码发送至电信运营商）；\n'
                  '（3）法律法规要求或政府机关依法要求。',
            ),
            _Section(
              title: '五、AI 数据处理',
              content: '5.1 本服务使用 AI 模型处理您的对话内容以生成回复。AI 处理过程遵循数据最小化原则，仅传输必要的上下文信息。\n\n'
                  '5.2 我们不会将您的个人对话数据用于 AI 模型的训练。\n\n'
                  '5.3 您的对话数据仅存储在您的个人账号下，其他用户无法访问。',
            ),
            _Section(
              title: '六、您的权利',
              content: '您对个人信息享有以下权利：\n'
                  '（1）访问权：您可以查看您的账号信息和使用数据；\n'
                  '（2）更正权：您可以更正不准确的个人信息；\n'
                  '（3）删除权：您可以删除部分或全部数据；\n'
                  '（4）注销权：您可以申请注销账号，我们将删除您的全部数据。\n\n'
                  '如需行使上述权利，请通过应用内反馈功能联系我们。',
            ),
            _Section(
              title: '七、政策更新',
              content: '7.1 我们可能适时更新本隐私政策，更新后的版本将在应用内公布。\n\n'
                  '7.2 重大变更我们将通过应用内通知或短信方式告知您。\n\n'
                  '7.3 继续使用本服务即表示您同意更新后的隐私政策。',
            ),
            _Section(
              title: '八、联系我们',
              content: '如您对本隐私政策有任何疑问或建议，请通过应用内反馈功能联系我们。',
            ),
            SizedBox(height: 48),
          ],
        ),
      ),
    );
  }
}

class _Section extends StatelessWidget {
  final String title;
  final String content;

  const _Section({required this.title, required this.content});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: const TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.w600,
              color: AppColors.textPrimary,

            ),
          ),
          const SizedBox(height: 8),
          Text(
            content,
            style: const TextStyle(
              fontSize: 14,
              color: AppColors.textSecondary,
              height: 1.6,

            ),
          ),
        ],
      ),
    );
  }
}
