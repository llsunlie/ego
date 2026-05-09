class OnboardingDiary {
  final String date;
  final String text;
  const OnboardingDiary({required this.date, required this.text});
}

class OnboardingGroup {
  final List<OnboardingDiary> diary;
  final OnboardingDiary diary2;
  final String insightFull;
  const OnboardingGroup({
    required this.diary,
    required this.diary2,
    required this.insightFull,
  });
}

const onboardingFeelings = [
  '最近还不错，但总有点隐隐的不踏实',
  '有点累，一种说不太清楚的累',
  '正在经历一些选择，还没想明白',
  '想记住最近某个好的瞬间',
];

const onboardingData = [
  OnboardingGroup(
    diary: [
      OnboardingDiary(
        date: '43 天前',
        text: '最近好像什么都顺。工作没出错，周末也有安排。但心里有一小块地方不太踏实，像在等另一只鞋掉下来。',
      ),
      OnboardingDiary(
        date: '55 天前',
        text: '今天在沙发上躺着，阳光正好，突然想：这种平静是真的吗？还是我只是暂时不去想那些事了？',
      ),
      OnboardingDiary(
        date: '38 天前',
        text: '好几天没emo了。但我发现自己在\'观察\'自己是不是真的好了。这本身就有点奇怪吧。',
      ),
    ],
    diary2: OnboardingDiary(
      date: '22 天前',
      text: '朋友说我最近状态不错。我说是啊。但我没说的是——我有点怕它走。',
    ),
    insightFull:
        '你选了"还不错但不踏实"——22天前你也写过类似的话："我有点怕它走。"\n\n你好像对"好"这件事不太信任。也许让你不安的不是"会不会变坏"，是你还不习惯就这样好着……',
  ),
  OnboardingGroup(
    diary: [
      OnboardingDiary(
        date: '62 天前',
        text: '又躺了一下午。不是身体累，是脑子不想动了。好像所有事情都跟我没关系，但我又不能真的什么都不做。',
      ),
      OnboardingDiary(
        date: '45 天前',
        text: '今天闹钟响了三次才起来。不是困，是不想面对今天。也没什么特别的事，就是不想开始。',
      ),
      OnboardingDiary(
        date: '71 天前',
        text: '别人问我最近怎么样，我说\'还行\'。其实不还行。但我也不知道哪里不行。',
      ),
    ],
    diary2: OnboardingDiary(
      date: '31 天前',
      text: '每天都在忙，但回头看什么都没留下。好像我只是在维持什么，不是在走向什么。',
    ),
    insightFull:
        '你选了"一种说不太清楚的累"——31天前你也写过："只是在维持，不是在走向什么。"\n\n你的累好像不是身体的、也不是某件事造成的。它更像一种"我在这里但我不知道为什么在这里"的感觉……',
  ),
  OnboardingGroup(
    diary: [
      OnboardingDiary(
        date: '50 天前',
        text: '跟他说了。说完之后好像松了，又好像更紧了。话已经出去了，收不回来了。但我不后悔说。',
      ),
      OnboardingDiary(
        date: '36 天前',
        text: '投了那个岗位。不是怕被拒，是怕被接受——万一真的要去了，我准备好了吗？',
      ),
      OnboardingDiary(
        date: '62 天前',
        text: '终于把那个计划列出来了。但写完有一种\'再也找不到借口拖着了\'的感觉。有点怕。',
      ),
    ],
    diary2: OnboardingDiary(
      date: '28 天前',
      text: '做决定之前想了很久。做完之后发现——原来让我纠结的不是\'选哪个\'，是\'要不要选\'本身。',
    ),
    insightFull:
        '你选了"正在经历选择"——28天前你也写过："让我纠结的不是选哪个，是要不要选本身。"\n\n你好像做完决定之后，不是在等结果，是在等自己的感觉告诉你"对了"。也许你已经知道答案了，只是还不敢信……',
  ),
  OnboardingGroup(
    diary: [
      OnboardingDiary(
        date: '47 天前',
        text: '中午吃到了很好吃的面。坐在窗边，外面在下雨。那一刻觉得什么都刚刚好。',
      ),
      OnboardingDiary(
        date: '33 天前',
        text: '跟朋友笑了一整个下午。回来路上一个人走，居然没有那种\'开心完就空\'的感觉。今天是被填满了的。',
      ),
      OnboardingDiary(
        date: '58 天前',
        text: '忽然发现我好久没哭了。不是忍着，是真的好久没有那么难受了。我想把这个记下来。',
      ),
    ],
    diary2: OnboardingDiary(
      date: '20 天前',
      text: '今天没什么事发生。但就是想来写一句。以前，我总是难受才会写东西。',
    ),
    insightFull:
        '你选了"想记住某个好的瞬间"——20天前你也写过："以前总是难受才写东西。"\n\n你正在学一件新的事：不只在痛苦的时候才看见自己，也在好的时候。这本身就是一种你可能还没注意到的变化……',
  ),
];

const onboardingPreviewLines = [
  '刚才和你说话的，是一个还不完全是你的 ego。',
  '如果你从今天开始把想法留在这里——',
  '几天后，你会开始看见自己反复出现的某些念头。',
  '再过一阵，你可以和过去不同时刻的自己对话——',
  '那些散落的话会彼此连接，让你看见自己正在经历的变化。',
  '某一天你回头看，会发现——',
  '你比自己以为的，更了解自己。',
];
