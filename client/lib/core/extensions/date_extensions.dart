extension DateExtensions on DateTime {
  String get monthLabel {
    const months = [
      '1月', '2月', '3月', '4月', '5月', '6月',
      '7月', '8月', '9月', '10月', '11月', '12月',
    ];
    return '$year年${months[month - 1]}';
  }

  String get dayLabel => '${month}月${day}日';

  String get weekdayLabel {
    const days = ['一', '二', '三', '四', '五', '六', '日'];
    return '星期${days[weekday - 1]}';
  }
}
