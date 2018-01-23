# concourse-preflight-snack

### A Concourse Pipeline Pre-Processor

Create templates within Concourse pipelines.

Given a pipeline with `pp_layouts:` defined, and a resource with `pp_tags:` specified:

This is a POC, has no tests, should not be used by anyone, etc.

```yaml
pp_layouts:
- name: go-tests                        # descriptive name
  type: job                             # type of this layout (job, resource)
  tags: [go]                            # tags this layout applies to
  groups: [all, test]                   # groups the rendered job should be added to
  template:                             # content of the template; `{{.Name}}` is the tagged resource's name
    name: "{{.Name}}-test-go1.9"
    plan:
    - get: "{{.Name}}"
      passed: ["push-{{.Name}}-image"]
      trigger: true
    - get: tasks
    - task: test
      file: tasks/test.yml

resources:
- name: repo1
  type: git
  pp_tags: [go]                         # list of tags that matches one or more layout above
  source:
    uri: git@example.com:org/repo1.git
    branch: master
    private_key: ((git_private_key))

- name: repo2
  type: git
  pp_tags: [go]
  source:
    uri: git@example.com:org/repo1.git
    branch: master
    private_key: ((git_private_key))

jobs: []                                # jobs will be appended to
groups: []                              # groups will be appended to
```

```
$ fly format-pipeline -c <(cat pipeline.yml | pp)
groups:
- name: all
  jobs:
  - repo1-test-go1.9
  - repo2-test-go1.9
- name: test
  jobs:
  - repo1-test-go1.9
  - repo2-test-go1.9
resources:
- name: repo1
  type: git
  source:
    branch: master
    private_key: ((git_private_key))
    uri: git@example.com:org/repo1.git
- name: repo2
  type: git
  source:
    branch: master
    private_key: ((git_private_key))
    uri: git@example.com:org/repo1.git
resource_types: []
jobs:
- name: repo1-test-go1.9
  plan:
  - get: repo1
    trigger: true
  - get: tasks
  - task: test
    file: tasks/test.yml
- name: repo2-test-go1.9
  plan:
  - get: repo2
    trigger: true
  - get: tasks
  - task: test
    file: tasks/test.yml

```

_"pp_layouts" is a reference to P.P. Layouts from [The Three Stigmata of Palmer Eldritch](https://en.wikipedia.org/wiki/The_Three_Stigmata_of_Palmer_Eldritch)._
