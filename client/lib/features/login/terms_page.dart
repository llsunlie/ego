import 'package:flutter/material.dart';
import '../../core/theme/colors.dart';

class TermsPage extends StatelessWidget {
  const TermsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('服务条款'),
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
              '服务条款',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
                fontFamily: 'NotoSansSC',
              ),
            ),
            SizedBox(height: 8),
            Text(
              '更新日期：2026年6月9日',
              style: TextStyle(fontSize: 13, color: AppColors.textSecondary, fontFamily: 'NotoSansSC'),
            ),
            SizedBox(height: 16),
            _Section(
              title: '一、服务说明',
              content: 'EGO 是一款基于人工智能的个人成长与自我探索应用（以下简称"本服务"）。本服务通过对话、记录、反思等功能，帮助用户更好地认识自己、管理情绪、追踪成长轨迹。',
            ),
            _Section(
              title: '二、用户注册与账号',
              content: '2.1 您在使用本服务前需要注册账号，注册时需提供手机号码。\n\n'
                  '2.2 您应当提供真实、准确的注册信息，并妥善保管账号和密码。因您保管不善导致的账号被盗用或损失，由您自行承担。\n\n'
                  '2.3 您不得将账号转让、出借或授权他人使用。一个手机号仅可注册一个账号。',
            ),
            _Section(
              title: '三、用户行为规范',
              content: '3.1 您在使用本服务过程中，应遵守中华人民共和国相关法律法规。\n\n'
                  '3.2 您不得利用本服务从事以下活动：\n'
                  '（1）发布、传播违法或不良信息；\n'
                  '（2）干扰本服务的正常运行；\n'
                  '（3）利用技术手段破解、反向工程本服务；\n'
                  '（4）其他违反法律法规或侵犯他人合法权益的行为。\n\n'
                  '3.3 如发现用户存在违规行为，我们有权暂停或终止向您提供服务。',
            ),
            _Section(
              title: '四、知识产权',
              content: '4.1 本服务的所有内容，包括但不限于文字、图片、软件、界面设计等，其知识产权归 EGO 所有或已获得合法授权。\n\n'
                  '4.2 您在使用本服务过程中产生的个人数据归您所有。您授予我们在提供服务所必需的范围内使用这些数据的权利。\n\n'
                  '4.3 未经明确授权，您不得复制、修改、传播本服务的任何内容。',
            ),
            _Section(
              title: '五、免责声明',
              content: '5.1 本服务提供的 AI 对话内容仅供参考，不构成任何医疗、心理或法律建议。如有心理健康问题，请寻求专业帮助。\n\n'
                  '5.2 我们致力于提供稳定、安全的服务，但不对因不可抗力、系统维护、网络故障等原因导致的服务中断承担责任。\n\n'
                  '5.3 我们有权在必要时修改本服务条款，修改后的条款将在应用内公布。继续使用本服务即表示您同意修改后的条款。',
            ),
            _Section(
              title: '六、终止服务',
              content: '6.1 您可随时停止使用本服务。如需注销账号，请联系我们。\n\n'
                  '6.2 如您违反本服务条款，我们有权暂停或终止向您提供服务，并保留追究法律责任的权利。',
            ),
            _Section(
              title: '七、法律适用与争议解决',
              content: '7.1 本条款的订立、执行和解释及争议的解决均适用中华人民共和国法律。\n\n'
                  '7.2 因本条款引起的争议，双方应友好协商解决；协商不成的，任何一方均可向有管辖权的人民法院提起诉讼。',
            ),
            _Section(
              title: '八、联系我们',
              content: '如您对本服务条款有任何疑问，请通过应用内反馈功能联系我们。',
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
              fontFamily: 'NotoSansSC',
            ),
          ),
          const SizedBox(height: 8),
          Text(
            content,
            style: const TextStyle(
              fontSize: 14,
              color: AppColors.textSecondary,
              height: 1.6,
              fontFamily: 'NotoSansSC',
            ),
          ),
        ],
      ),
    );
  }
}
