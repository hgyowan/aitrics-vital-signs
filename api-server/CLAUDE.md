# 개발 배경
- 해당 개발건은 aitrics 의 과제 전형을 위함입니다.
- ai-history 경로 하위에 AITRICS.md 파일에 해당 과제애대한 내용이 담겨있습니다. 이를 먼저 꼼꼼하게 참고하세요.

# 개발 규칙 Summary
- 해당 프로젝트는 layered + ddd lite 아키텍처 형태로 구성 되었습니다. 
- 각 디렉토리 및 파일 구조를 확인하고 하위 CLAUDE.md 파일을 꼭 확인하여 작업하세요.
- 하나의 작업 단위당 현재까지 한 대화 내용은 ai-history 디렉토리 하위에 md 파일로 대화 했던 내용 및 개발 사항에대해 건건이 순서대로 요약하여 알아보기 쉽게 파일명을 작성하여 생성합니다.
  - 파일명은 1단계-{내용}.md 와같이 생성하세요.
- 네이밍 규칙, 구조, 설계 규칙은 기존 코드 스타일과 100% 동일한 패턴을 유지하도록 하세요.
- git 과 관련된 행위는 진행하지 않습니다.

# 개발 순서 (중요)
- /api-server/domain 하위에 도메인 구성을 위한 repository, service, controller interface 를 선언하세요. (entity 는 이미 정의되어 있습니다.)
- /api-server/app 하위에 위에서 정의한 interface 의 구현체를 각 layer 에 작성하세요.
- interface 작성후 항상 각 layer 의 mocking 을 진행해 주세요.
  - 각 layer 파일 상단에 generate 코드가 존재합니다. 이를 참고하세요
  ```go
  //go:generate mockgen -source=controller.go -destination=../mock/mock_controller.go -package=mock
  ```
- 추가로 생성되는 코드에 대해 모든 layer 에 필수적으로 test code 를 작성하세요. 기존 test 코드 작성 규칙을 참고하세요.
  - 주입받는 대상은 모두 mocking 되어 어떠한 환경에서도 테스트 가능하도록 구성 되어어야 합니다.
  - 테이블 드리븐 테스트 방식으로 진행하세요.
  - 테스트는 테스트 대상의 커버리지가 최대한 70% 가 넘도록 작성하세요.
  - 테스트 코드는 비즈니스 로직을 이해할 수 있는 문서와같이 느껴져야 합니다. 단순히 동작만 하는 테스트 코드를 작성하지 마세요.
- /api-server/app/router 경로에 요구사항에 맞는 api 를 추가하세요.