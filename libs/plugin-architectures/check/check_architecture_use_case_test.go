package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"
)

// tree writes files, given as path then content pairs.
func tree(t *testing.T, pairs ...string) string {
	t.Helper()
	root := t.TempDir()
	for i := 0; i < len(pairs); i += 2 {
		p := filepath.Join(root, filepath.FromSlash(pairs[i]))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(pairs[i+1]), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func run(t *testing.T, arch, lang string, pairs ...string) Report {
	t.Helper()
	rs, ok := rules.For(arch)
	if !ok {
		t.Fatalf("no rules for %s", arch)
	}
	r, err := Check(tree(t, pairs...), lang, rs)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func wantProblems(t *testing.T, r Report, want ...string) {
	t.Helper()
	got := strings.Join(r.Problems, "\n")
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in:\n%s", w, got)
		}
	}
	if len(r.Problems) != len(want) {
		t.Errorf("want %d problems, got:\n%s", len(want), got)
	}
}

const goMod = "module example.com/shop\n\ngo 1.26\n"

func TestHexagonalGo(t *testing.T) {
	r := run(t, "hexagonal", "go",
		"go.mod", goMod,
		"domain/order.go", "package domain\n",
		"application/place_order.go", "package application\nimport (\n\t\"example.com/shop/domain\"\n\t\"fmt\"\n\t\"github.com/lib/pq\"\n)\n",
		"domain/bad.go", "package domain\nimport \"example.com/shop/adapters/postgres\"\n",
		"adapters/postgres/store.go", "package postgres\nimport \"example.com/shop/domain\"\n",
		"adapters/http/handler.go", "package http\nimport (\n\t\"example.com/shop/application\"\n\t\"example.com/shop/adapters/postgres\"\n)\n",
		"adapters/http/handler_test.go", "package http\nimport \"example.com/shop/adapters/postgres\"\n",
		"adapters/postgres/base.go", "package postgres\n",
		"adapters/shared.go", "package adapters\n",
		"adapters/http/uses_shared.go", "package http\nimport \"example.com/shop/adapters\"\n",
		"main.go", "package main\nimport _ \"example.com/shop/adapters/http\"\n",
	)
	wantProblems(t, r,
		"domain/bad.go - imports example.com/shop/adapters/postgres: core code must not depend on adapters code",
		"adapters/http/handler.go - imports example.com/shop/adapters/postgres: adapters http must not depend on adapters postgres",
	)
}

func TestHexagonalSeparatesInboundAndOutboundFolders(t *testing.T) {
	r := run(t, "hexagonal", "go",
		"go.mod", goMod,
		"adapters/inbound/rest/handler.go", "package rest\nimport \"example.com/shop/adapters/outbound/sql\"\n",
		"adapters/outbound/sql/store.go", "package sql\n",
	)
	wantProblems(t, r, "adapters rest must not depend on adapters sql")
}

func TestCleanNode(t *testing.T) {
	r := run(t, "clean", "node",
		"src/entities/order.ts", "export class Order {}\n",
		"src/use-cases/place-order.ts", "import { Order } from '../entities/order'\nimport express from 'express'\n",
		"src/use-cases/leaky.ts", "import { Db } from '../frameworks/db'\nimport { x } from '../interface-adapters/presenter.js'\n",
		"src/entities/leaky.ts", "import { PlaceOrder } from '../use-cases/place-order'\n",
		"src/interface-adapters/controller.ts", "const uc = require('../use-cases/place-order')\nimport { Db } from '@/frameworks/db'\n",
		"src/frameworks/db.ts", "import { Order } from '../entities/order'\n",
		"src/entities/order.spec.ts", "import { Db } from '../frameworks/db'\n",
	)
	wantProblems(t, r,
		"src/use-cases/leaky.ts - imports ../frameworks/db: use-cases code must not depend on frameworks code",
		"src/use-cases/leaky.ts - imports ../interface-adapters/presenter.js: use-cases code must not depend on interface-adapters code",
		"src/entities/leaky.ts - imports ../use-cases/place-order: entities code must not depend on use-cases code",
		"src/interface-adapters/controller.ts - imports @/frameworks/db",
	)
}

func TestCleanPythonSeesRelativeAndPackageImports(t *testing.T) {
	r := run(t, "clean", "python",
		"shop/__init__.py", "",
		"shop/entities/__init__.py", "",
		"shop/entities/order.py", "from shop.use_cases import place_order\nimport shop.frameworks.db\nimport requests\nfrom ..use_cases.place import x\nfrom .. import frameworks\n",
		"shop/use_cases/place_order.py", "from shop.entities.order import Order\nfrom django.views import View\n",
		"shop/frameworks/db.py", "",
		"tests/test_order.py", "from shop.frameworks import db\n",
	)
	wantProblems(t, r,
		"shop/entities/order.py - imports shop.use_cases: entities code must not depend on use-cases code",
		"shop/entities/order.py - imports shop.frameworks.db: entities code must not depend on frameworks code",
		"shop/entities/order.py - imports ..use_cases.place: entities code must not depend on use-cases code",
		"shop/entities/order.py - imports .. import frameworks: entities code must not depend on frameworks code",
	)
}

func TestHexagonalJava(t *testing.T) {
	r := run(t, "hexagonal", "jvm",
		"src/main/java/com/acme/domain/Order.java", "package com.acme.domain;\nimport com.acme.adapters.jpa.OrderEntity;\nimport java.util.List;\nimport org.springframework.stereotype.Service;\n",
		"src/main/java/com/acme/adapters/jpa/OrderEntity.java", "package com.acme.adapters.jpa;\nimport com.acme.domain.Order;\n",
		"src/test/java/com/acme/domain/OrderTest.java", "package com.acme.domain;\nimport com.acme.adapters.jpa.OrderEntity;\n",
	)
	wantProblems(t, r, "Order.java - imports com.acme.adapters.jpa.OrderEntity: core code must not depend on adapters code")
}

func TestHexagonalCSharpDartRustPhpRubyCpp(t *testing.T) {
	cases := []struct {
		lang string
		bad  []string
		good []string
	}{
		{"dotnet", []string{"Domain/Order.cs", "namespace Shop.Domain;\nusing Shop.Adapters.Sql;\nusing System;\n"}, []string{"Adapters/Sql/Store.cs", "namespace Shop.Adapters.Sql;\nusing Shop.Domain;\n"}},
		{"dart", []string{"lib/src/domain/order.dart", "import 'package:shop/src/adapters/sql/store.dart';\nimport 'dart:io';\n"}, []string{"lib/src/adapters/sql/store.dart", "import '../../domain/order.dart';\n"}},
		{"rust", []string{"src/domain/order.rs", "use crate::adapters::sql::Store;\n"}, []string{"src/adapters/sql.rs", "use crate::domain::order;\nuse super::sql;\n"}},
		{"php", []string{"src/Domain/Order.php", "<?php\nnamespace Shop\\Domain;\nuse Shop\\Adapters\\Sql\\Store;\nuse DateTime;\n"}, []string{"src/Adapters/Sql/Store.php", "<?php\nnamespace Shop\\Adapters\\Sql;\nuse Shop\\Domain\\Order;\n"}},
		{"ruby", []string{"lib/shop/domain/order.rb", "require_relative '../adapters/store'\nrequire 'json'\n"}, []string{"lib/shop/adapters/store.rb", "require_relative '../domain/order'\n"}},
		{"cpp", []string{"src/domain/order.cpp", "#include \"adapters/store.h\"\n#include <vector>\n"}, []string{"src/adapters/store.cpp", "#include \"domain/order.h\"\n"}},
	}
	for _, c := range cases {
		t.Run(c.lang, func(t *testing.T) {
			pairs := append([]string{}, c.bad...)
			pairs = append(pairs, c.good...)
			if c.lang == "dart" {
				pairs = append(pairs, "pubspec.yaml", "name: shop\n")
			}
			r := run(t, "hexagonal", c.lang, pairs...)
			wantProblems(t, r, c.bad[0]+" - imports ")
		})
	}
}

func TestDDDKeepsLayersInOrderAndContextsApart(t *testing.T) {
	r := run(t, "ddd", "go",
		"go.mod", goMod,
		"sales/domain/order.go", "package domain\nimport \"example.com/shop/sales/infrastructure\"\n",
		"sales/application/place.go", "package application\nimport (\n\t\"example.com/shop/sales/domain\"\n\t\"example.com/shop/billing/domain\"\n\t\"example.com/shop/billing\"\n)\n",
		"sales/infrastructure/db.go", "package infrastructure\nimport (\n\t\"example.com/shop/sales/application\"\n\t\"example.com/shop/sales/domain\"\n)\n",
		"billing/domain/invoice.go", "package domain\n",
		"billing/api.go", "package billing\n",
		"billing/infrastructure/db.go", "package infrastructure\nimport \"example.com/shop/billing/application\"\n",
		"billing/application/pay.go", "package application\nimport \"example.com/shop/billing/domain\"\n",
	)
	wantProblems(t, r,
		"sales/domain/order.go - imports example.com/shop/sales/infrastructure: domain code must not depend on infrastructure code",
		"sales/application/place.go - imports example.com/shop/billing/domain: reaches into the domain layer of the billing context",
	)
}

func TestDDDInJavaTellsContextsByPackage(t *testing.T) {
	r := run(t, "ddd", "jvm",
		"src/main/java/com/acme/orders/application/Place.java", "package com.acme.orders.application;\nimport com.acme.orders.domain.Order;\nimport com.acme.billing.infrastructure.Db;\n",
		"src/main/java/com/acme/orders/domain/Order.java", "package com.acme.orders.domain;\n",
		"src/main/java/com/acme/billing/infrastructure/Db.java", "package com.acme.billing.infrastructure;\n",
	)
	wantProblems(t, r, "Place.java - imports com.acme.billing.infrastructure.Db: reaches into the infrastructure layer of the billing context")
}

func TestModularMonolithGo(t *testing.T) {
	r := run(t, "modular-monolith", "go",
		"go.mod", goMod,
		"modules/orders/orders.go", "package orders\nimport \"example.com/shop/modules/orders/store\"\n",
		"modules/orders/api/api.go", "package api\n",
		"modules/orders/store/store.go", "package store\n",
		"modules/billing/pay.go", "package billing\nimport (\n\t\"example.com/shop/modules/orders\"\n\t\"example.com/shop/modules/orders/api\"\n\t\"example.com/shop/modules/orders/store\"\n)\n",
		"main.go", "package main\nimport (\n\t\"example.com/shop/modules/billing\"\n\t\"example.com/shop/modules/orders/store\"\n)\n",
	)
	wantProblems(t, r,
		"modules/billing/pay.go - imports example.com/shop/modules/orders/store: reaches into the internals of the orders module",
		"main.go - imports example.com/shop/modules/orders/store: reaches into the internals of the orders module",
	)
}

func TestModularMonolithNode(t *testing.T) {
	r := run(t, "modular-monolith", "node",
		"src/modules/orders/index.ts", "export * from './internal/place'\n",
		"src/modules/orders/internal/place.ts", "export const x = 1\n",
		"src/modules/billing/pay.ts", "import { a } from '../orders'\nimport { b } from '../orders/index'\nimport { c } from '../orders/internal/place'\n",
	)
	wantProblems(t, r, "src/modules/billing/pay.ts - imports ../orders/internal/place: reaches into the internals of the orders module")
}

func TestAProjectWithoutLayerFoldersIsReportedNotFailed(t *testing.T) {
	r := run(t, "hexagonal", "go", "go.mod", goMod, "main.go", "package main\n", "store/store.go", "package store\n")
	if !r.OK() || len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "none of its rules were applied") {
		t.Fatalf("%+v", r)
	}
}

func TestAProjectWithoutSourcesIsFine(t *testing.T) {
	r := run(t, "clean", "go", "go.mod", goMod)
	if !r.OK() || len(r.Warnings) != 0 {
		t.Fatalf("%+v", r)
	}
}

func TestAnUnknownLanguageIsRefused(t *testing.T) {
	rs, _ := rules.For("clean")
	if _, err := Check(t.TempDir(), "cobol", rs); err == nil || !strings.Contains(err.Error(), "knows") {
		t.Fatalf("%v", err)
	}
}

func TestLayerFoldersListsWhatIsNamedForALayer(t *testing.T) {
	rs, _ := rules.For("hexagonal")
	dir := tree(t, "go.mod", goMod, "domain/a.go", "package domain\n", "internal/adapters/http/h.go", "package http\n")
	got, err := LayerFolders(dir, "go", rs)
	if err != nil || len(got) != 2 {
		t.Fatalf("%v %v", got, err)
	}
}
